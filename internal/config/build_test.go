package config

import (
	"strings"
	"testing"
)

func TestBuildConfigValidateTargets(t *testing.T) {
	tests := []struct {
		name    string
		targets []BuildTarget
		want    []string
	}{
		{
			name: "accepts legacy unnamed targets",
			targets: []BuildTarget{
				{Platform: "linux-x64", Cmd: "make linux"},
				{Platform: "windows-x64", Cmd: "make windows"},
			},
		},
		{
			name: "requires command",
			targets: []BuildTarget{
				{Name: "linux-release", Platform: "linux-x64"},
			},
			want: []string{"build.targets[0] (linux-release) has no cmd"},
		},
		{
			name: "rejects duplicate names case insensitively",
			targets: []BuildTarget{
				{Name: "Release", Cmd: "make linux"},
				{Name: "release", Cmd: "make windows"},
			},
			want: []string{"build.targets[1] name \"release\" duplicates build.targets[0]"},
		},
		{
			name: "checks every target",
			targets: []BuildTarget{
				{Name: "debug", Cmd: "make debug"},
				{Name: "release", Cmd: "   "},
				{Name: "RELEASE", Cmd: "make release"},
			},
			want: []string{
				"build.targets[1] (release) has no cmd",
				"build.targets[2] name \"RELEASE\" duplicates build.targets[1]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := (BuildConfig{Targets: tt.targets}).ValidateTargets()
			if len(errs) != len(tt.want) {
				t.Fatalf("got %d errors %v, want %d", len(errs), errs, len(tt.want))
			}
			for i, want := range tt.want {
				if !strings.Contains(errs[i], want) {
					t.Errorf("error %q does not contain %q", errs[i], want)
				}
			}
		})
	}
}
