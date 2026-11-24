package o3de

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRegisteredReadsO3DEManifestArray(t *testing.T) {
	manifest := filepath.Join(t.TempDir(), "o3de_manifest.json")
	t.Setenv("O3DE_MANIFEST_PATH", manifest)
	content := `{"engines":["/opt/o3de","/work/another-engine"]}`
	if err := os.WriteFile(manifest, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	entries, err := Registered("engines")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %#v", entries)
	}
	found := map[string]bool{}
	for _, entry := range entries {
		found[entry.Path] = true
	}
	if !found["/opt/o3de"] || !found["/work/another-engine"] {
		t.Fatalf("unexpected entries: %#v", entries)
	}
}

func TestRegisteredRejectsUnknownKind(t *testing.T) {
	if _, err := Registered("unknown"); err == nil {
		t.Fatal("expected unknown kind error")
	}
}

func TestFindEngineRequiresManifest(t *testing.T) {
	dir := t.TempDir()
	if _, err := FindEngine(dir); err == nil {
		t.Fatal("expected missing engine.json error")
	}
}

func TestFindEngineReadsEngineManifest(t *testing.T) {
	dir := t.TempDir()
	content := `{"engine_name":"TestEngine","version":"24.09"}`
	if err := os.WriteFile(filepath.Join(dir, "engine.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := FindEngine(dir)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name != "TestEngine" || entry.Version != "24.09" {
		t.Fatalf("unexpected engine: %#v", entry)
	}
}

func TestCLIInvocationUsesCommandInterpreterForWindowsBatchFiles(t *testing.T) {
	for _, path := range []string{`C:\Program Files\O3DE\scripts\o3de.bat`, `C:\Tools\o3de.cmd`} {
		executable, args, err := CLIInvocation("windows", path, []string{"register", "--engine-path", `C:\My Engine`})
		if err != nil {
			t.Fatal(err)
		}
		if executable != "cmd.exe" {
			t.Fatalf("expected cmd.exe for %q, got %q", path, executable)
		}
		want := []string{"/d", "/v:off", "/s", "/c", `""` + path + `" "register" "--engine-path" "C:\My Engine""`}
		if !reflect.DeepEqual(args, want) {
			t.Fatalf("unexpected command args: got %#v, want %#v", args, want)
		}
	}
}

func TestCLIInvocationRunsNonBatchScriptsDirectly(t *testing.T) {
	for _, goos := range []string{"linux", "darwin", "windows"} {
		executable, args, err := CLIInvocation(goos, "/opt/o3de/scripts/o3de.sh", []string{"find", "engines"})
		if err != nil {
			t.Fatal(err)
		}
		if executable != "/opt/o3de/scripts/o3de.sh" {
			t.Fatalf("unexpected executable for %s: %q", goos, executable)
		}
		if !reflect.DeepEqual(args, []string{"find", "engines"}) {
			t.Fatalf("unexpected args for %s: %#v", goos, args)
		}
	}
}

func TestCLIInvocationQuotesWindowsBatchArguments(t *testing.T) {
	executable, args, err := CLIInvocation("windows", `C:\Program Files\O3DE\scripts\o3de.bat`, []string{"register", "--project-name", "Game & Tools"})
	if err != nil {
		t.Fatal(err)
	}
	if executable != "cmd.exe" {
		t.Fatalf("expected cmd.exe, got %q", executable)
	}
	want := []string{"/d", "/v:off", "/s", "/c", `""C:\Program Files\O3DE\scripts\o3de.bat" "register" "--project-name" "Game & Tools""`}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected command args: got %#v, want %#v", args, want)
	}
}

func TestCLIInvocationRejectsBatchArgumentsThatCannotBeSafelyQuoted(t *testing.T) {
	for _, value := range []string{"100%ready", `quoted"name`, "line\nbreak", "line\rbreak"} {
		t.Run(value, func(t *testing.T) {
			_, _, err := CLIInvocation("windows", `C:\O3DE\scripts\o3de.bat`, []string{"register", "--project-name", value})
			if err == nil {
				t.Fatalf("CLIInvocation accepted unsafe batch argument %q", value)
			}
		})
	}
}

func TestCLIInvocationRejectsBatchScriptPathWithPercent(t *testing.T) {
	_, _, err := CLIInvocation("windows", `C:\O3DE%BUILD%\scripts\o3de.bat`, nil)
	if err == nil {
		t.Fatal("CLIInvocation accepted a batch script path containing percent expansion")
	}
}
