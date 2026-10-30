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
