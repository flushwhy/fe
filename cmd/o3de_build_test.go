package cmd

import (
	"reflect"
	"testing"
)

func TestO3DEBuildArgs(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		targets []string
		want    []string
	}{
		{
			name: "single config default",
			want: []string{"--build", "build", "--parallel"},
		},
		{
			name:   "multi config",
			config: "Profile",
			want:   []string{"--build", "build/linux", "--parallel", "--config", "Profile"},
		},
		{
			name:    "multiple targets",
			targets: []string{"Game", "Game.Editor"},
			want:    []string{"--build", "build", "--parallel", "--target", "Game", "Game.Editor"},
		},
		{
			name:    "config and targets",
			config:  "Release",
			targets: []string{"MyGem"},
			want:    []string{"--build", "out", "--parallel", "--config", "Release", "--target", "MyGem"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buildDir := "build"
			if test.name == "multi config" { buildDir = "build/linux" }
			if test.name == "config and targets" { buildDir = "out" }
			got := o3deBuildArgs(buildDir, test.config, test.targets)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("o3deBuildArgs() = %#v, want %#v", got, test.want)
			}
		})
	}
}
