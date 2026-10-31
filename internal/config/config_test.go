package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadBuildTargetsSupportsLegacyAndStructuredYAML(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	for _, name := range []string{"ITCHIO_USERNAME", "ITCHIO_GAME", "FE_USERVERSION", "FE_BUILD_DIR"} {
		t.Setenv(name, "")
	}

	path := filepath.Join(t.TempDir(), ".fe.yaml")
	yaml := `build:
  targets:
    - platform: linux-x64
      cmd: "gcc -o builds/game src/main.c"
    - platform: host
      cmd: "/opt/Build Tools/compiler"
      args:
        - "-o"
        - "build output/game"
        - "source files/main.c"
`
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}

	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Build.Targets) != 2 {
		t.Fatalf("got %d build targets, want 2", len(cfg.Build.Targets))
	}

	legacy := cfg.Build.Targets[0]
	if legacy.Cmd != "gcc -o builds/game src/main.c" || legacy.Args != nil {
		t.Fatalf("legacy target changed during config load: %#v", legacy)
	}

	structured := cfg.Build.Targets[1]
	if structured.Cmd != "/opt/Build Tools/compiler" {
		t.Fatalf("structured command = %q", structured.Cmd)
	}
	wantArgs := []string{"-o", "build output/game", "source files/main.c"}
	if !reflect.DeepEqual(structured.Args, wantArgs) {
		t.Fatalf("structured args = %#v, want %#v", structured.Args, wantArgs)
	}
}
