// Package scaffold creates new projects from remote templates.
package scaffold

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/flushwhy/fe/internal/ui"
)

type Lang string

const (
	LangOdin Lang = "odin"
	LangC    Lang = "c"
	LangCpp  Lang = "cpp"
	LangZig  Lang = "zig"
	LangGo   Lang = "go"
)

var SupportedLangs = []Lang{LangOdin, LangC, LangCpp, LangZig, LangGo}

type Options struct {
	Lang      Lang
	Name      string // project directory name
	Dir       string // base directory to create project in (default: cwd)
	Module    string // Go module path
	RemoteURL string
	TemplateID  string // named template from flushwhy/fe-templates
	templateFS  fs.FS  // optional template filesystem, used by tests
	GoVersion string
}

type TemplateData struct {
	Name      string
	Lang      string
	Module    string
	GoVersion string
}

// projectRoot returns the absolute path where the project will be created.
func (o Options) projectRoot() (string, error) {
	if o.Name == "" || o.Name == "." || o.Name == ".." || filepath.IsAbs(o.Name) || filepath.Base(filepath.Clean(o.Name)) != o.Name || strings.ContainsAny(o.Name, `/\\`) {
		return "", fmt.Errorf("project name must be a single directory name, not a path: %q", o.Name)
	}
	base := o.Dir
	if base == "" {
		var err error
		base, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("could not get working directory: %w", err)
		}
	}
	return filepath.Join(base, o.Name), nil
}

func Init(opts Options) error {
	explicitTemplate := opts.TemplateID != ""
	if !explicitTemplate {
		if err := validateLang(opts.Lang); err != nil {
			return err
		}
		opts.TemplateID = defaultTemplateID(opts.Lang)
	}
	if !templateIDPattern.MatchString(opts.TemplateID) {
		return fmt.Errorf("invalid template ID %q (use lowercase words separated by hyphens)", opts.TemplateID)
	}
	if opts.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if opts.GoVersion == "" {
		opts.GoVersion = goVersion()
	}
	if opts.Module == "" {
		opts.Module = opts.Name
	}

	root, err := opts.projectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(root); !os.IsNotExist(err) {
		return fmt.Errorf("directory %q already exists", opts.Name)
	}

	data := TemplateData{
		Name:      opts.Name,
		Lang:      string(opts.Lang),
		Module:    opts.Module,
		GoVersion: opts.GoVersion,
	}

	label := string(opts.Lang)
	if explicitTemplate {
		label = opts.TemplateID
	}
	ui.Title("%s  Initialising %s project: %s", ui.IconRocket, label, opts.Name)

	var src fs.FS

	err = ui.RunSpinner("Resolving templates", func() error {
		if opts.templateFS != nil {
			src = opts.templateFS
			return nil
		}
		if opts.RemoteURL != "" && !explicitTemplate {
			var fetchErr error
			src, fetchErr = fetchRemote(opts.RemoteURL, string(opts.Lang))
			if fetchErr != nil {
				return fmt.Errorf("fetching configured remote template: %w", fetchErr)
			}
			return nil
		}
		cachedPath, err := ResolveTemplate(opts.TemplateID, false)
		if err != nil {
			return fmt.Errorf("could not resolve template %q; check your internet connection and try again: %w", opts.TemplateID, err)
		}
		src = os.DirFS(cachedPath)
		return nil
	})
	if err != nil {
		return err
	}

	if err := os.Mkdir(root, 0750); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("directory %q already exists", opts.Name)
		}
		return fmt.Errorf("creating project directory %q: %w", opts.Name, err)
	}

	err = ui.RunSpinner("Writing project files", func() error {
		return writeTemplate(src, root, data)
	})
	if err != nil {
		return err
	}

	err = ui.RunSpinner("Finalising project", func() error {
		return postInit(opts, root)
	})
	if err != nil {
		return err
	}

	fmt.Println()
	ui.OK("Project %q ready", opts.Name)
	if opts.Lang != "" {
		printNextSteps(opts.Lang, opts.Name)
	} else {
		ui.Hint(fmt.Sprintf("cd %s", opts.Name), "Review the template README for build instructions.")
	}
	return nil
}

func writeTemplate(src fs.FS, destRoot string, data TemplateData) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !fs.ValidPath(path) || path == "." {
			if path == "." && d.IsDir() {
				info, err := os.Lstat(destRoot)
				if err != nil {
					return err
				}
				if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
					return fmt.Errorf("project destination is not a real directory")
				}
				return nil
			}
			return fmt.Errorf("unsafe template path %q", path)
		}
		dest := filepath.Join(destRoot, filepath.FromSlash(path))
		rel, err := filepath.Rel(destRoot, dest)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("template path escapes project directory: %q", path)
		}
		if d.IsDir() {
			return os.MkdirAll(dest, 0750)
		}
		content, err := fs.ReadFile(src, path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}
		destPath := strings.TrimSuffix(dest, ".tmpl")
		if filepath.Base(destPath) == "gitignore" {
			destPath = filepath.Join(filepath.Dir(destPath), ".gitignore")
		}
		rendered, err := renderTemplate(string(content), data)
		if err != nil {
			return fmt.Errorf("rendering template %s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
			return err
		}
		file, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
		if err != nil {
			return fmt.Errorf("creating template file %s: %w", path, err)
		}
		if _, err := io.WriteString(file, rendered); err != nil {
			_ = file.Close()
			_ = os.Remove(destPath)
			return fmt.Errorf("writing template file %s: %w", path, err)
		}
		if err := file.Close(); err != nil {
			_ = os.Remove(destPath)
			return fmt.Errorf("closing template file %s: %w", path, err)
		}
		return nil
	})
}

func renderTemplate(tmpl string, data TemplateData) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func postInit(opts Options, root string) error {
	switch opts.Lang {
	case LangC, LangCpp:
		stub := filepath.Join(root, "compile_commands.json")
		if err := os.WriteFile(stub, []byte("[]\n"), 0640); err != nil {
			return fmt.Errorf("could not write compile_commands.json stub: %w", err)
		}
	}
	return nil
}

func printNextSteps(lang Lang, name string) {
	switch lang {
	case LangOdin:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add ols            — wire up OLS for Neovim",
			"odin run . -file      — run the project",
		)
	case LangC:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add clangd         — wire up clangd for Neovim",
			"bear -- make          — generate compile_commands.json",
		)
	case LangCpp:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add clangd         — wire up clangd for Neovim",
			"cmake -B build -DCMAKE_EXPORT_COMPILE_COMMANDS=ON",
		)
	case LangZig:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add zls            — wire up ZLS for Neovim",
			"zig build run",
		)
	case LangGo:
		ui.Hint(
			fmt.Sprintf("cd %s", name),
			"fe add gopls          — wire up gopls for Neovim",
			"go run .",
		)
	}
}

func fetchRemote(baseURL, lang string) (fs.FS, error) {
	url := strings.TrimRight(baseURL, "/") + "/" + lang + ".tar.gz"
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote returned %d for %s", resp.StatusCode, url)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxTemplateArchiveBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if len(data) > maxTemplateArchiveBytes {
		return nil, fmt.Errorf("remote template archive exceeds %d MiB limit", maxTemplateArchiveBytes>>20)
	}
	return extractTarGz(data)
}

func defaultTemplateID(lang Lang) string {
	switch lang {
	case LangC:
		return "c-make"
	case LangCpp:
		return "cpp-cmake"
	case LangGo:
		return "go-basic"
	case LangOdin:
		return "odin-basic"
	case LangZig:
		return "zig-build"
	default:
		return ""
	}
}

func validateLang(lang Lang) error {
	for _, l := range SupportedLangs {
		if l == lang {
			return nil
		}
	}
	names := make([]string, len(SupportedLangs))
	for i, l := range SupportedLangs {
		names[i] = string(l)
	}
	return fmt.Errorf("unsupported language %q — supported: %s", lang, strings.Join(names, ", "))
}

func goVersion() string {
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}
