package update

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestChocolateyMetadata(t *testing.T) {
	id, version, err := chocolateyMetadata([]byte(`<package><metadata><id>fe</id><version>0.8.0</version></metadata></package>`))
	if err != nil || id != "fe" || version != "0.8.0" {
		t.Fatalf("got %q %q %v", id, version, err)
	}
}

func TestChocolateyMetadataRejectsMalformedXML(t *testing.T) {
	_, _, err := chocolateyMetadata([]byte("<package>"))
	if err == nil {
		t.Fatal("expected malformed nuspec to fail")
	}
}

func TestPackageOwnsPath(t *testing.T) {
	path := filepath.Clean("/usr/bin/fe")
	if !packageOwnsPath("fe: "+path+"\n", path) {
		t.Fatal("expected Fe package ownership")
	}
	if packageOwnsPath("other: "+path+"\n", path) {
		t.Fatal("accepted another package as Fe")
	}
	if packageOwnsPath("fe: /usr/local/bin/fe\n", path) {
		t.Fatal("accepted a different executable path")
	}
}

func TestDetectRequiresPackageOwnership(t *testing.T) {
	d := Detector{
		GOOS: "linux", Executable: "/usr/local/bin/fe",
		LookPath: func(name string) (string, error) {
			if name == "apt-get" || name == "dpkg-query" { return "/usr/bin/"+name, nil }
			return "", errors.New("not found")
		},
		Run: func(_ context.Context, name string, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "-S" { return []byte("other: /usr/local/bin/fe\n"), nil }
			return []byte("0.8.0"), nil
		},
	}
	if _, err := d.Detect(context.Background(), APT); err == nil {
		t.Fatal("expected unowned executable to be rejected")
	}
}


func TestDetectChocolateyUsesInstalledPackageMetadata(t *testing.T) {
	root := t.TempDir()
	payload := filepath.Join(root, "lib", "fe", "tools", "fe.exe")
	shim := filepath.Join(root, "bin", "fe.exe")
	if err := os.MkdirAll(filepath.Dir(payload), 0755); err != nil { t.Fatal(err) }
	if err := os.MkdirAll(filepath.Dir(shim), 0755); err != nil { t.Fatal(err) }
	if err := os.WriteFile(payload, []byte("binary"), 0600); err != nil { t.Fatal(err) }
	nuspec := filepath.Join(root, "lib", "fe", "fe.nuspec")
	if err := os.WriteFile(nuspec, []byte("<package><metadata><id>fe</id><version>0.8.0</version></metadata></package>"), 0600); err != nil { t.Fatal(err) }

	d := Detector{
		GOOS: "windows", Executable: shim,
		Env: func(name string) string { if name == "ChocolateyInstall" { return root }; return "" },
	}
	got, err := d.Detect(context.Background(), Chocolatey)
	if err != nil { t.Fatal(err) }
	if got.Method != Chocolatey || got.PackageVersion != "0.8.0" { t.Fatalf("unexpected detection: %#v", got) }

	d.Executable = filepath.Join(root, "other", "fe.exe")
	if _, err := d.Detect(context.Background(), Chocolatey); err == nil {
		t.Fatal("accepted an executable not owned by the Chocolatey package")
	}
}

func TestDetectAutoDoesNotGuessDirectBinary(t *testing.T) {
	d := Detector{
		GOOS: "linux", Executable: "/usr/local/bin/fe",
		LookPath: func(name string) (string, error) {
			if name == "apt-get" || name == "dpkg-query" { return "/usr/bin/"+name, nil }
			return "", errors.New("not found")
		},
		Run: func(context.Context, string, ...string) ([]byte, error) {
			return nil, errors.New("not package-owned")
		},
	}
	if got, err := d.Detect(context.Background(), Auto); err != nil || got.Method != "" {
		t.Fatalf("unknown installation should remain unclassified: %#v, %v", got, err)
	}
}


func TestPackageCandidateParsers(t *testing.T) {
	choco, err := parseChocolateyCandidate([]byte("fe|0.4.0\nfe|0.8.0\nother|9.9.9\n"))
	if err != nil || choco != "0.8.0" { t.Fatalf("Chocolatey candidate = %q, %v", choco, err) }

	brew, err := parseHomebrewCandidate([]byte(`{"formulae":[{"versions":{"stable":"0.8.0"}}]}`))
	if err != nil || brew != "0.8.0" { t.Fatalf("Homebrew candidate = %q, %v", brew, err) }

	apt, err := parseAPTCandidate([]byte("fe:\n  Installed: 0.6.0\n  Candidate: 0.8.0\n"))
	if err != nil || apt != "0.8.0" { t.Fatalf("APT candidate = %q, %v", apt, err) }
	if _, err := parseAPTCandidate([]byte("fe:\n  Installed: 0.6.0\n  Candidate: (none)\n")); err == nil {
		t.Fatal("expected missing APT candidate to fail")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct { a, b string; want int }{
		{"0.8.0", "0.7.9", 1},
		{"v0.8.0", "0.8.0", 0},
		{"0.8.0", "0.8.1", -1},
		{"0.8.0-rc.1", "0.8.0", -1},
		{"0.8.0", "0.8.0-rc.1", 1},
	}
	for _, test := range tests {
		if got := CompareVersions(test.a, test.b); got != test.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", test.a, test.b, got, test.want)
		}
	}
}

func TestUpdateCommand(t *testing.T) {
	tests := []struct { method Method; name string; args []string }{
		{Chocolatey, "choco", []string{"upgrade", "fe", "-y", "--no-progress"}},
		{Homebrew, "brew", []string{"upgrade", "fe"}},
		{APT, "sudo", []string{"apt-get", "install", "--only-upgrade", "fe"}},
	}
	for _, test := range tests {
		got, err := UpdateCommand(test.method, "stable")
		if err != nil { t.Fatal(err) }
		if got.Name != test.name || !reflect.DeepEqual(got.Args, test.args) {
			t.Errorf("%s command = %#v", test.method, got)
		}
	}
	if _, err := UpdateCommand(Homebrew, "rolling"); err == nil {
		t.Fatal("Homebrew must not claim to publish the rolling channel")
	}
}

func TestCompareAPTVersions(t *testing.T) {
	tests := []struct {
		installed string
		release   string
		want      int
	}{
		{"0.8.0-1", "0.8.0", 0},
		{"1:0.8.0-2", "0.8.0", 0},
		{"0.7.9-1", "0.8.0", -1},
		{"0.8.0~rc1-1", "0.8.0", -1},
	}
	for _, test := range tests {
		if got := ComparePackageVersions(APT, test.installed, test.release); got != test.want {
			t.Errorf("ComparePackageVersions(APT, %q, %q) = %d, want %d", test.installed, test.release, got, test.want)
		}
	}
}

func TestDetectAutoRejectsAmbiguousPackageOwnership(t *testing.T) {
	executable := "/usr/bin/fe"
	d := Detector{
		GOOS:       "linux",
		Executable: executable,
		LookPath: func(name string) (string, error) {
			switch name {
			case "apt-get", "dpkg-query":
				return "/usr/bin/" + name, nil
			case "brew":
				return "/home/linuxbrew/.linuxbrew/bin/brew", nil
			default:
				return "", errors.New("not found")
			}
		},
		Run: func(_ context.Context, name string, args ...string) ([]byte, error) {
			switch {
			case strings.HasSuffix(name, "dpkg-query") && args[0] == "-S":
				return []byte("fe: " + executable + "\n"), nil
			case strings.HasSuffix(name, "dpkg-query"):
				return []byte("ii|0.8.0"), nil
			case strings.HasSuffix(name, "brew") && args[1] == "--versions":
				return []byte("fe 0.8.0"), nil
			case strings.HasSuffix(name, "brew"):
				return []byte(executable + "\n"), nil
			default:
				return nil, errors.New("unexpected command")
			}
		},
	}

	if _, err := d.Detect(context.Background(), Auto); err == nil || !strings.Contains(err.Error(), "multiple package managers") {
		t.Fatalf("expected ambiguous ownership to be rejected, got %v", err)
	}
}
