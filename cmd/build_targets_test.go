package cmd

import (
	"strings"
	"testing"

	"github.com/flushwhy/fe/internal/config"
)

func TestSelectBuildTargets(t *testing.T) {
	configured := []config.BuildTarget{
		{Name: "linux-debug", Platform: "linux-x64", Cmd: "make debug"},
		{Name: "linux-release", Platform: "linux-x64", Cmd: "make release"},
		{Name: "windows-release", Platform: "windows-x64", Cmd: "make windows"},
	}

	tests := []struct {
		name     string
		target   string
		platform string
		want     []string
		wantErr  string
	}{
		{name: "all targets", want: []string{"linux-debug", "linux-release", "windows-release"}},
		{name: "named target", target: "linux-release", want: []string{"linux-release"}},
		{name: "platform selects every match", platform: "linux-x64", want: []string{"linux-debug", "linux-release"}},
		{name: "unknown target", target: "missing", wantErr: "no build target named"},
		{name: "unknown platform", platform: "freebsd-x64", wantErr: "no target found for platform"},
		{name: "conflicting selectors", target: "linux-release", platform: "linux-x64", wantErr: "cannot be used together"},
		{name: "duplicate named target", target: "same", wantErr: "is not unique"},
	}

	duplicateNames := append(append([]config.BuildTarget{}, configured...), config.BuildTarget{
		Name: "same", Platform: "linux-x64", Cmd: "make one",
	}, config.BuildTarget{
		Name: "same", Platform: "linux-x64", Cmd: "make two",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets := configured
			if tt.name == "duplicate named target" {
				targets = duplicateNames
			}
			got, err := selectBuildTargets(targets, tt.target, tt.platform)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d targets, want %d: %#v", len(got), len(tt.want), got)
			}
			for i, want := range tt.want {
				if got[i].Name != want {
					t.Errorf("target %d = %q, want %q", i, got[i].Name, want)
				}
			}
		})
	}
}
