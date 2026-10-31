package build

import (
	"reflect"
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func TestPlanResolvesCommandsAndFiltersPlatform(t *testing.T) {
	cfg := config.BuildConfig{Targets: []config.BuildTarget{
		{
			Platform: "linux-x64",
			Cmd:      "go",
			Args:     []string{"build", "-ldflags=-X main.version=1.2.0", "./..."},
			Output:   "builds/linux-x64/game",
		},
		{Platform: "windows-x64", Cmd: "gcc -o game.exe main.c"},
	}}

	plan, err := Plan(cfg, Options{Platform: "LINUX-X64"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 {
		t.Fatalf("got %d planned targets, want 1", len(plan))
	}
	got := plan[0]
	if got.Command != "go" || got.Platform != "linux-x64" {
		t.Fatalf("unexpected planned target: %#v", got)
	}
	wantArgs := []string{"build", "-ldflags=-X main.version=1.2.0", "./..."}
	if !reflect.DeepEqual(got.Args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", got.Args, wantArgs)
	}
	if got.Output != "builds/linux-x64/game" {
		t.Fatalf("output = %q", got.Output)
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
	if len(plan) != 1 {
		t.Fatalf("got %d targets, want 1", len(plan))
	}
	if plan[0].Command != "/opt/Build Tools/compiler" {
		t.Fatalf("command = %q", plan[0].Command)
	}
	wantArgs := []string{"-o", "build output/game", "source files/main.c", "--define=NAME=value with spaces"}
	if !reflect.DeepEqual(plan[0].Args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", plan[0].Args, wantArgs)
	}
}

func TestPlanRetainsLegacyCommandParsing(t *testing.T) {
	cfg := config.BuildConfig{Targets: []config.BuildTarget{{
		Platform: "linux-x64",
		Cmd:      "gcc -o game main.c",
	}}}

	plan, err := Plan(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Command != "gcc" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	wantArgs := []string{"-o", "game", "main.c"}
	if !reflect.DeepEqual(plan[0].Args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", plan[0].Args, wantArgs)
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
}

func TestPlanWithSuggestionsRejectsCrossPlatformSelection(t *testing.T) {
	suggestions := []config.BuildTarget{{Cmd: "go", Args: []string{"build", "./..."}}}
	_, err := PlanWithSuggestions(config.BuildConfig{}, Options{Platform: "windows-x64"}, suggestions)
	if err == nil || !strings.Contains(err.Error(), "current host only") {
		t.Fatalf("expected host-only suggestion error, got %v", err)
	}
}

func TestCheckToolsReportsMissingExecutable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	err := CheckTools([]PlannedTarget{{Platform: "host", Command: "fe-tool-that-does-not-exist"}})
	if err == nil || !strings.Contains(err.Error(), "was not found in PATH") {
		t.Fatalf("expected missing tool error, got %v", err)
	}
}
