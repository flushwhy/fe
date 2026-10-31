package build

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func TestPlanResolvesCommandsAndFiltersPlatform(t *testing.T) {
	cfg := config.BuildConfig{Targets: []config.BuildTarget{
		{Platform: "linux-x64", Cmd: "go", Args: []string{"build", "-ldflags=-X main.version=1.2.0", "./..."}, Output: "builds/linux-x64/game"},
		{Platform: "windows-x64", Cmd: "gcc -o game.exe main.c"},
	}}
	plan, err := Plan(cfg, Options{Platform: "LINUX-X64"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Command != "go" || plan[0].Platform != "linux-x64" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if want := []string{"build", "-ldflags=-X main.version=1.2.0", "./..."}; !reflect.DeepEqual(plan[0].Args, want) {
		t.Fatalf("args = %#v, want %#v", plan[0].Args, want)
	}
	if plan[0].Output != "builds/linux-x64/game" {
		t.Fatalf("output = %q", plan[0].Output)
	}
}

func TestPlanPreservesStructuredCommandAndArguments(t *testing.T) {
	cfg := config.BuildConfig{Targets: []config.BuildTarget{{
		Platform: "host",
		Cmd:      "/opt/Build Tools/compiler",
		Args:     []string{"-o", "build output/game", "source files/main.c", "--define=NAME=value with spaces"},
	}}}
	plan, err := Plan(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Command != "/opt/Build Tools/compiler" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	want := []string{"-o", "build output/game", "source files/main.c", "--define=NAME=value with spaces"}
	if !reflect.DeepEqual(plan[0].Args, want) {
		t.Fatalf("args = %#v, want %#v", plan[0].Args, want)
	}
}

func TestPlanRetainsLegacyCommandParsing(t *testing.T) {
	cfg := config.BuildConfig{Targets: []config.BuildTarget{{Platform: "linux-x64", Cmd: "gcc -o game main.c"}}}
	plan, err := Plan(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Command != "gcc" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if want := []string{"-o", "game", "main.c"}; !reflect.DeepEqual(plan[0].Args, want) {
		t.Fatalf("args = %#v, want %#v", plan[0].Args, want)
	}
}

func TestPlanValidatesAllTargetsBeforeExecution(t *testing.T) {
	cfg := config.BuildConfig{Targets: []config.BuildTarget{
		{Platform: "linux-x64", Cmd: "go build ./..."},
		{Platform: "windows-x64"},
	}}
	plan, err := Plan(cfg, Options{})
	if err == nil || !strings.Contains(err.Error(), "windows-x64") {
		t.Fatalf("expected missing command error for second target, got plan=%#v err=%v", plan, err)
	}
	if plan != nil {
		t.Fatalf("expected no partial plan, got %#v", plan)
	}
}

func TestPlanRejectsMissingTargetsAndUnknownPlatform(t *testing.T) {
	if _, err := Plan(config.BuildConfig{}, Options{}); err == nil {
		t.Fatal("expected missing targets error")
	}
	cfg := config.BuildConfig{Targets: []config.BuildTarget{{Platform: "linux-x64", Cmd: "go"}}}
	if _, err := Plan(cfg, Options{Platform: "freebsd-x64"}); err == nil {
		t.Fatal("expected unknown platform error")
	}
}

func TestPlanWithSuggestionsUsesDetectedTargetsOnlyWhenUnconfigured(t *testing.T) {
	suggestions := []config.BuildTarget{{Cmd: "go", Args: []string{"build", "./..."}}}
	plan, err := PlanWithSuggestions(config.BuildConfig{}, Options{}, suggestions)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Platform != "host" || plan[0].Command != "go" {
		t.Fatalf("unexpected detected plan: %#v", plan)
	}

	explicit := config.BuildConfig{Targets: []config.BuildTarget{{Platform: "linux-x64", Cmd: "gcc"}}}
	plan, err = PlanWithSuggestions(explicit, Options{}, suggestions)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Command != "gcc" || plan[0].Platform != "linux-x64" {
		t.Fatalf("explicit config was not authoritative: %#v", plan)
	}

	_, err = PlanWithSuggestions(explicit, Options{Platform: "windows-x64"}, suggestions)
	if err == nil || !strings.Contains(err.Error(), "windows-x64") {
		t.Fatalf("explicit targets must not fall back to suggestions: %v", err)
	}
}

func TestPlanWithSuggestionsRejectsCrossPlatformSelection(t *testing.T) {
	suggestions := []config.BuildTarget{{Cmd: "go", Args: []string{"build", "./..."}}}
	_, err := PlanWithSuggestions(config.BuildConfig{}, Options{Platform: "windows-x64"}, suggestions)
	if err == nil || !strings.Contains(err.Error(), "current host only") {
		t.Fatalf("expected host-only suggestion error, got %v", err)
	}
}

func TestCheckToolsValidatesEveryTargetBeforeExecution(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())

	// An absolute executable path remains valid even when PATH is empty.
	plan := []PlannedTarget{
		{Platform: "first", Command: executable},
		{Platform: "second", Command: "fe-missing-build-tool"},
	}
	err = CheckTools(plan)
	if err == nil || !strings.Contains(err.Error(), "second") || !strings.Contains(err.Error(), "fe-missing-build-tool") {
		t.Fatalf("expected preflight to report the missing second tool, got %v", err)
	}
}
