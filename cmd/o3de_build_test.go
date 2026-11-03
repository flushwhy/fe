package cmd

import (
	"reflect"
	"testing"
)

func TestO3DEBuildArgs(t *testing.T) {
	tests := []struct {
		name, buildDir, config string
		targets, want          []string
	}{
		{"defaults", "build", "", nil, []string{"--build", "build", "--parallel"}},
		{"multi-config", "build/linux", "Profile", nil, []string{"--build", "build/linux", "--parallel", "--config", "Profile"}},
		{"multiple-targets", "build", "", []string{"Game", "Game.Editor"}, []string{"--build", "build", "--parallel", "--target", "Game", "Game.Editor"}},
		{"config-and-target", "out", "Release", []string{"MyGem"}, []string{"--build", "out", "--parallel", "--config", "Release", "--target", "MyGem"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := o3deBuildArgs(test.buildDir, test.config, test.targets)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("o3deBuildArgs() = %#v, want %#v", got, test.want)
			}
		})
	}
}
