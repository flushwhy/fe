package update

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Method string

const (
	Auto Method = "auto"
	Chocolatey Method = "chocolatey"
	Homebrew Method = "homebrew"
	APT Method = "apt"
	Direct Method = "direct"
)

type Installation struct {
	Method Method
	PackageVersion string
	Executable string
}

type Command struct { Name string; Args []string }
type Release struct { Version, Commit, URL string }
type Runner func(context.Context, string, ...string) ([]byte, error)

type Detector struct {
	GOOS string
	Executable string
	Env func(string) string
	LookPath func(string) (string, error)
	Run Runner
}

func Detect(ctx context.Context, requested Method) (Installation, error) {
	exe, err := os.Executable()
	if err != nil { return Installation{}, fmt.Errorf("resolve running executable: %w", err) }
	d := Detector{GOOS: runtime.GOOS, Executable: exe, Env: os.Getenv, LookPath: exec.LookPath}
	d.Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).Output()
	}
	return d.Detect(ctx, requested)
}

func (d Detector) Detect(ctx context.Context, requested Method) (Installation, error) {
	if d.Env == nil { d.Env = func(string) string { return "" } }
	if d.LookPath == nil { d.LookPath = func(string) (string, error) { return "", errors.New("not found") } }
	if d.Run == nil { d.Run = func(context.Context, string, ...string) ([]byte, error) { return nil, errors.New("runner unavailable") } }
	if requested == Direct { return Installation{Method: Direct}, nil }
	if requested != Auto {
		install, err := d.detectMethod(ctx, requested)
		if err != nil { return Installation{}, fmt.Errorf("cannot verify %s ownership: %w", requested, err) }
		return install, nil
	}
	var found []Installation
	for _, method := range methodsForOS(d.GOOS) {
		if install, err := d.detectMethod(ctx, method); err == nil { found = append(found, install) }
	}
	switch len(found) {
	case 0: return Installation{}, nil
	case 1: return found[0], nil
	default:
		names := make([]string, len(found))
		for i, install := range found { names[i] = string(install.Method) }
		return Installation{}, fmt.Errorf("installation metadata matches multiple package managers (%s); rerun with --method", strings.Join(names, ", "))
	}
}

func methodsForOS(goos string) []Method {
	switch goos {
	case "windows": return []Method{Chocolatey}
	case "darwin": return []Method{Homebrew}
	case "linux": return []Method{APT, Homebrew}
	default: return nil
	}
}

func (d Detector) detectMethod(ctx context.Context, method Method) (Installation, error) {
	switch method {
	case Chocolatey:
		if d.GOOS != "windows" { return Installation{}, errors.New("Chocolatey is only supported on Windows") }
		root := d.Env("ChocolateyInstall")
		if root == "" {
			if programData := d.Env("ProgramData"); programData != "" { root = filepath.Join(programData, "chocolatey") }
		}
		if root == "" { return Installation{}, errors.New("ChocolateyInstall metadata is unavailable") }
		nuspec := filepath.Join(root, "lib", "fe", "fe.nuspec")
		payload := filepath.Join(root, "lib", "fe", "tools", "fe.exe")
		shim := filepath.Join(root, "bin", "fe.exe")
		data, err := os.ReadFile(nuspec)
		if err != nil { return Installation{}, err }
		id, version, err := chocolateyMetadata(data)
		if err != nil { return Installation{}, err }
		if !strings.EqualFold(id, "fe") || version == "" { return Installation{}, errors.New("Chocolatey metadata does not identify Fe") }
		if _, err := os.Stat(payload); err != nil { return Installation{}, errors.New("Chocolatey Fe package payload is missing") }
		if !samePath(d.Executable, shim, d.GOOS) && !samePath(d.Executable, payload, d.GOOS) { return Installation{}, errors.New("running executable is not owned by the Chocolatey Fe package") }
		return Installation{Method: Chocolatey, PackageVersion: version, Executable: d.Executable}, nil

	case Homebrew:
		if d.GOOS != "darwin" && d.GOOS != "linux" { return Installation{}, errors.New("Homebrew is only supported on macOS and Linux") }
		brew, err := d.LookPath("brew")
		if err != nil { return Installation{}, errors.New("brew is not available") }
		versions, err := d.Run(ctx, brew, "list", "--versions", "flushwhy/tap/fe")
		if err != nil { return Installation{}, errors.New("Homebrew does not report Fe as installed") }
		fields := strings.Fields(string(versions))
		if len(fields) < 2 || fields[0] != "fe" { return Installation{}, errors.New("Homebrew Fe installation metadata is invalid") }
		files, err := d.Run(ctx, brew, "list", "--verbose", "flushwhy/tap/fe")
		if err != nil || !containsPath(string(files), d.Executable, d.GOOS) { return Installation{}, errors.New("running executable is not listed in Homebrew's Fe formula metadata") }
		return Installation{Method: Homebrew, PackageVersion: fields[1], Executable: d.Executable}, nil

	case APT:
		if d.GOOS != "linux" { return Installation{}, errors.New("APT is only supported on Linux") }
		if _, err := d.LookPath("apt-get"); err != nil { return Installation{}, errors.New("apt-get is not available") }
		version, err := d.debPackageVersion(ctx)
		if err != nil { return Installation{}, err }
		return Installation{Method: APT, PackageVersion: version, Executable: d.Executable}, nil

	default: return Installation{}, fmt.Errorf("unsupported update method %q", method)
	}
}

func chocolateyMetadata(data []byte) (string, string, error) {
	var nuspec struct {
		Metadata struct {
			ID      string `xml:"id"`
			Version string `xml:"version"`
		} `xml:"metadata"`
	}
	if err := xml.Unmarshal(data, &nuspec); err != nil {
		return "", "", fmt.Errorf("parse Chocolatey nuspec: %w", err)
	}
	return strings.TrimSpace(nuspec.Metadata.ID), strings.TrimSpace(nuspec.Metadata.Version), nil
}

func (d Detector) debPackageVersion(ctx context.Context) (string, error) {
	query, err := d.LookPath("dpkg-query")
	if err != nil { return "", errors.New("dpkg-query is not available") }
	paths := []string{d.Executable}
	if resolved, err := filepath.EvalSymlinks(d.Executable); err == nil && resolved != d.Executable { paths = append(paths, resolved) }
	owned := false
	for _, path := range paths {
		out, err := d.Run(ctx, query, "-S", path)
		if err == nil && packageOwnsPath(string(out), path) { owned = true; break }
	}
	if !owned { return "", errors.New("dpkg does not report ownership of the running executable") }
	format := "-f=$" + "{db:Status-Abbrev}|$" + "{Version}"
	out, err := d.Run(ctx, query, "-W", format, "fe")
	if err != nil { return "", errors.New("installed dpkg package Fe could not be queried") }
	fields := strings.SplitN(strings.TrimSpace(string(out)), "|", 2)
	if len(fields) != 2 || !strings.HasPrefix(fields[0], "ii") { return "", errors.New("dpkg package Fe is not in the installed state") }
	return strings.TrimSpace(fields[1]), nil
}

func packageOwnsPath(output, path string) bool {
	for _, line := range strings.Split(output, "\n") {
		left, right, ok := strings.Cut(strings.TrimSpace(line), ": ")
		if ok && (left == "fe" || strings.HasPrefix(left, "fe:")) && filepath.Clean(right) == filepath.Clean(path) { return true }
	}
	return false
}

func samePath(a, b, goos string) bool {
	a, _ = filepath.Abs(a); b, _ = filepath.Abs(b)
	if goos == "windows" { return strings.EqualFold(filepath.Clean(a), filepath.Clean(b)) }
	return filepath.Clean(a) == filepath.Clean(b)
}

func containsPath(list, executable, goos string) bool {
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil { resolved = executable }
	for _, line := range strings.Split(list, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && (samePath(line, executable, goos) || samePath(line, resolved, goos)) { return true }
	}
	return false
}

func Latest(ctx context.Context, channel string) (Release, error) {
	if channel == "rolling" {
		var result struct {
			SHA string `json:"sha"`
		}
		if err := getJSON(ctx, "https://api.github.com/repos/flushwhy/fe/commits/rolling", &result); err != nil {
			return Release{}, err
		}
		if result.SHA == "" {
			return Release{}, errors.New("GitHub returned no rolling commit")
		}
		return Release{Version: "rolling", Commit: result.SHA, URL: "https://github.com/flushwhy/fe/releases/tag/rolling"}, nil
	}
	if channel != "stable" {
		return Release{}, fmt.Errorf("unknown update channel %q (expected stable or rolling)", channel)
	}

	var result struct {
		TagName string `json:"tag_name"`
		URL     string `json:"html_url"`
	}
	if err := getJSON(ctx, "https://api.github.com/repos/flushwhy/fe/releases/latest", &result); err != nil {
		return Release{}, err
	}
	if result.TagName == "" {
		return Release{}, errors.New("GitHub returned a release without a tag")
	}
	return Release{
		Version: strings.TrimPrefix(strings.TrimPrefix(result.TagName, "fe-v"), "v"),
		URL:     result.URL,
	}, nil
}

func getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create GitHub release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("check GitHub releases: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub release check returned %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode GitHub release metadata: %w", err)
	}
	return nil
}

func Candidate(ctx context.Context, install Installation) (string, error) {
	var name string
	var args []string
	switch install.Method {
	case Chocolatey: name, args = "choco", []string{"search", "fe", "--exact", "--limit-output", "--no-progress"}
	case Homebrew: name, args = "brew", []string{"info", "--json=v2", "flushwhy/tap/fe"}
	case APT: name, args = "apt-cache", []string{"policy", "fe"}
		default: return "", fmt.Errorf("no package source query for %s", install.Method)
	}
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil { return "", fmt.Errorf("query %s package source: %w", install.Method, err) }
	switch install.Method {
	case Chocolatey:
		return parseChocolateyCandidate(out)
	case Homebrew:
		return parseHomebrewCandidate(out)
	case APT:
		return parseAPTCandidate(out)
	}

	return "", fmt.Errorf("no Fe package candidate is available from configured %s sources", install.Method)
}

func parseChocolateyCandidate(out []byte) (string, error) {
	candidate := ""
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Split(strings.TrimSpace(line), "|")
		if len(fields) >= 2 && strings.EqualFold(fields[0], "fe") && CompareVersions(fields[1], candidate) > 0 {
			candidate = fields[1]
		}
	}
	if candidate == "" { return "", errors.New("no Fe package candidate was found") }
	return candidate, nil
}

func parseHomebrewCandidate(out []byte) (string, error) {
	var info struct {
		Formulae []struct {
			FullName string `json:"full_name"`
			Versions struct {
				Stable string `json:"stable"`
			} `json:"versions"`
		} `json:"formulae"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return "", fmt.Errorf("parse Homebrew formula metadata: %w", err)
	}
	if len(info.Formulae) == 0 || info.Formulae[0].FullName != "flushwhy/tap/fe" || info.Formulae[0].Versions.Stable == "" {
		return "", errors.New("Homebrew metadata does not identify the Fe tap formula")
	}
	return info.Formulae[0].Versions.Stable, nil
}

func parseAPTCandidate(out []byte) (string, error) {
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Candidate:") {
			candidate := strings.TrimSpace(strings.TrimPrefix(line, "Candidate:"))
			if candidate != "(none)" && candidate != "" { return candidate, nil }
			break
		}
	}
	return "", errors.New("no Fe package candidate is available from configured APT sources")
}

func CompareVersions(a, b string) int {
	a = strings.TrimPrefix(strings.TrimSpace(a), "v")
	b = strings.TrimPrefix(strings.TrimSpace(b), "v")
	parse := func(value string) ([]int, string) {
		base, suffix, _ := strings.Cut(value, "-")
		base, _, _ = strings.Cut(base, "+")
		fields := strings.Split(base, ".")
		numbers := make([]int, 3)
		for i := 0; i < len(fields) && i < 3; i++ { numbers[i], _ = strconv.Atoi(fields[i]) }
		return numbers, suffix
	}
	av, apre := parse(a); bv, bpre := parse(b)
	for i := range av { if av[i] < bv[i] { return -1 }; if av[i] > bv[i] { return 1 } }
	if apre == bpre { return 0 }
	if apre == "" { return 1 }
	if bpre == "" { return -1 }
	return strings.Compare(apre, bpre)
}

// ComparePackageVersions accounts for Debian epochs and package revisions.
func ComparePackageVersions(method Method, installed, release string) int {
	if method == APT {
		installed = debianUpstreamVersion(installed)
		release = debianUpstreamVersion(release)
	}
	return CompareVersions(installed, release)
}

func debianUpstreamVersion(version string) string {
	if _, upstream, ok := strings.Cut(version, ":"); ok {
		version = upstream
	}
	if revision := strings.LastIndexByte(version, '-'); revision >= 0 {
		version = version[:revision]
	}
	return strings.Replace(version, "~", "-", 1)
}

func UpdateCommand(method Method, channel string) (Command, error) {
	if channel != "stable" { return Command{}, fmt.Errorf("%s does not publish Fe's rolling channel", method) }
	switch method {
	case Chocolatey: return Command{Name: "choco", Args: []string{"upgrade", "fe", "-y", "--no-progress"}}, nil
	case Homebrew: return Command{Name: "brew", Args: []string{"upgrade", "flushwhy/tap/fe"}}, nil
	case APT: return Command{Name: "sudo", Args: []string{"apt-get", "install", "--only-upgrade", "fe"}}, nil
		default: return Command{}, fmt.Errorf("updates through %s are not supported", method)
	}
}
