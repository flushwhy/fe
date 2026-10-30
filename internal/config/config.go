package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config is the single typed representation of .fe.yaml + env vars + flags.
// Priority: env vars > .fe.yaml > conventions > defaults
type Config struct {
	Itchio    ItchioConfig    `mapstructure:"itchio"`
	Butler    ButlerConfig    `mapstructure:"butler"`
	Pack      PackConfig      `mapstructure:"pack"`
	Transcode TranscodeConfig `mapstructure:"transcode"`
	Build     BuildConfig     `mapstructure:"build"`
	Godot     GodotConfig     `mapstructure:"godot"`
	Preflight PreflightConfig `mapstructure:"preflight"`
}

type ItchioConfig struct {
	Username string `mapstructure:"username"`
	Game     string `mapstructure:"game"`
}

type ButlerConfig struct {
	Directory   string `mapstructure:"directory"`
	Userversion string `mapstructure:"userversion"`
}

type PackConfig struct {
	Input  string               `mapstructure:"input"`
	Output string               `mapstructure:"output"`
	Sheets map[string]PackSheet `mapstructure:"sheets"`
}

// PackSheet configures one named spritesheet. Empty paths use conventions
// based on the pack input/output and the sheet name.
type PackSheet struct {
	Input  string `mapstructure:"input"`
	Output string `mapstructure:"output"`
}

type TranscodeConfig struct {
	InputFile       string `mapstructure:"inputFile"`
	OutputFile      string `mapstructure:"outputFile"`
	Codec           string `mapstructure:"codec"`
	Bitrate         string `mapstructure:"bitrate"`
	AudioChannels   string `mapstructure:"audioChannels"`
	VideoFrameRate  string `mapstructure:"videoFrameRate"`
	VideoResolution string `mapstructure:"videoResolution"`
	StartTime       string `mapstructure:"startTime"`
	EndTime         string `mapstructure:"endTime"`
}

type GodotConfig struct {
	Executable string `mapstructure:"executable"`
	OutputDir  string `mapstructure:"output_dir"`
}

type PreflightConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Detect  bool `mapstructure:"detect"`
	Checks  []PreflightCheck `mapstructure:"checks"`
}

type PreflightCheck struct {
	Name    string   `mapstructure:"name"`
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
	Stage   string   `mapstructure:"stage"`
	Fix     string   `mapstructure:"fix"`
}

type BuildConfig struct {
	Targets   []BuildTarget    `mapstructure:"targets"`
	Container ContainerBuild   `mapstructure:"container"`
}

type ContainerBuild struct {
	Image   string `mapstructure:"image"`
	Command string `mapstructure:"command"`
}

// BuildTarget represents a single platform build.
type BuildTarget struct {
	Platform string `mapstructure:"platform"`
	Cmd      string   `mapstructure:"cmd"`
	Args     []string `mapstructure:"args"`
	Output   string   `mapstructure:"output"`
}

// Load reads viper state into a typed Config, then overlays env vars.
func Load() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	applyEnvOverrides(&cfg)
	applyConventions(&cfg)
	if !viper.IsSet("preflight.detect") {
		cfg.Preflight.Detect = true
	}

	return &cfg, nil
}

// applyEnvOverrides maps well-known env vars onto the config struct.
// Env vars always win — CI secrets slot in here.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ITCHIO_USERNAME"); v != "" {
		cfg.Itchio.Username = v
	}
	if v := os.Getenv("ITCHIO_GAME"); v != "" {
		cfg.Itchio.Game = v
	}
	if v := os.Getenv("FE_USERVERSION"); v != "" {
		cfg.Butler.Userversion = v
	}
	if v := os.Getenv("FE_BUILD_DIR"); v != "" {
		cfg.Butler.Directory = v
	}
}

// applyConventions fills in missing values using project folder conventions.
// Only sets a value if it isn't already configured.
func applyConventions(cfg *Config) {
	if cfg.Butler.Directory == "" {
		cfg.Butler.Directory = "builds"
	}
	if cfg.Godot.Executable == "" {
		cfg.Godot.Executable = "godot"
	}
	if cfg.Godot.OutputDir == "" {
		cfg.Godot.OutputDir = cfg.Butler.Directory
	}
	if cfg.Pack.Input == "" {
		cfg.Pack.Input = "assets/sprites"
	}
	if cfg.Pack.Output == "" {
		cfg.Pack.Output = "assets/spritesheet.png"
	}
	if cfg.Transcode.OutputFile == "" {
		cfg.Transcode.OutputFile = "output.ogg"
	}
}

// Validate checks that required fields are present for the given command.
func (c *Config) Validate(command string) []string {
	var errs []string

	switch command {
	case "bmp", "release":
		if c.Itchio.Username == "" {
			errs = append(errs, "itchio.username is required (set in .fe.yaml or ITCHIO_USERNAME env var)")
		}
		if c.Itchio.Game == "" {
			errs = append(errs, "itchio.game is required (set in .fe.yaml or ITCHIO_GAME env var)")
		}
	case "transcode":
		if c.Transcode.InputFile == "" {
			errs = append(errs, "transcode.inputFile is required (set --inputFile flag)")
		}
	case "pack":
		if len(c.Pack.Sheets) == 0 && c.Pack.Input == "" {
			errs = append(errs, "pack.input is required when pack.sheets is not configured (set --input or pack.input in .fe.yaml)")
		}
	}

	return errs
}

// DefaultYAML is the content written by `fe init`.
const DefaultYAML = `# fe configuration
# Docs: https://github.com/flushwhy/fe

# Optional: pull init templates from a remote source instead of built-ins.
# templates:
#   source: "https://github.com/flushwhy/fe-templates"

itchio:
  username: ""   # or set ITCHIO_USERNAME env var
  game: ""       # or set ITCHIO_GAME env var

butler:
  directory: "builds"      # root folder containing platform subdirs
  userversion: ""          # optional — or set FE_USERVERSION env var

godot:
  executable: "godot"     # executable name or path
  output_dir: "builds"    # export output root; platform subfolders are added automatically

pack:
  input: "assets/sprites"
  output: "assets/spritesheet.png"
  # Optional: pack separate sprite groups instead of one combined sheet.
  # sheets:
  #   player:
  #     input: "assets/sprites/player"
  #     output: "assets/player.png"
  #   npcs:
  #     input: "assets/sprites/npcs"
  #     output: "assets/npcs.png"

transcode:
  codec: "libvorbis"
  bitrate: "128k"

build:
  # Optional: build inside a temporary Docker/Podman container.
  # Run with: fe build --container [--runtime docker|podman]
  # container:
  #   image: flushy/odin:nightly
  #   command: "odin build . -out:builds/game"
  targets:
    - platform: linux-x64
      cmd: gcc
      args: [-o, builds/linux-x64/game, src/main.c]
    - platform: windows-x64
      cmd: x86_64-w64-mingw32-gcc
      args: [-o, builds/windows-x64/game.exe, src/main.c]
    - platform: macos-arm64
      cmd: gcc
      args: [-o, builds/macos-arm64/game, src/main.c]
`

// SummaryLines returns a human-readable summary of active config values,
// masking empty fields so CI logs stay clean.
func (c *Config) SummaryLines() []string {
	lines := []string{
		fmt.Sprintf("  itch.io user : %s", maskEmpty(c.Itchio.Username)),
		fmt.Sprintf("  itch.io game : %s", maskEmpty(c.Itchio.Game)),
		fmt.Sprintf("  builds dir   : %s", c.Butler.Directory),
		fmt.Sprintf("  userversion  : %s", maskEmpty(c.Butler.Userversion)),
		fmt.Sprintf("  pack input   : %s", c.Pack.Input),
		fmt.Sprintf("  pack output  : %s", c.Pack.Output),
	}
	if len(c.Pack.Sheets) > 0 {
		lines = append(lines, fmt.Sprintf("  pack sheets  : %d named sheets", len(c.Pack.Sheets)))
	}

	if len(c.Build.Targets) > 0 {
		platforms := make([]string, len(c.Build.Targets))
		for i, t := range c.Build.Targets {
			platforms[i] = t.Platform
		}
		lines = append(lines, fmt.Sprintf("  build targets: %s", strings.Join(platforms, ", ")))
	}

	return lines
}

func maskEmpty(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}
