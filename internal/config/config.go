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

type BuildConfig struct {
	Targets []BuildTarget `mapstructure:"targets"`
}

// BuildTarget represents a single platform build.
type BuildTarget struct {
	Platform string `mapstructure:"platform"`
	Cmd      string `mapstructure:"cmd"`
	Output   string `mapstructure:"output"`
}

// Load reads viper state into a typed Config, then overlays env vars.
func Load() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	applyEnvOverrides(&cfg)
	applyConventions(&cfg)

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
		if c.Pack.Input == "" {
			errs = append(errs, "pack.input is required (set --input flag or pack.input in .fe.yaml)")
		}
	case "build":
		if len(c.Build.Targets) == 0 {
			errs = append(errs, "build.targets is empty — add at least one target in .fe.yaml")
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

pack:
  input: "assets/sprites"
  output: "assets/spritesheet.png"

transcode:
  codec: "libvorbis"
  bitrate: "128k"

build:
  targets:
    - platform: linux-x64
      cmd: "gcc -o builds/linux-x64/game src/main.c"
    - platform: windows-x64
      cmd: "x86_64-w64-mingw32-gcc -o builds/windows-x64/game.exe src/main.c"
    - platform: macos-arm64
      cmd: "gcc -o builds/macos-arm64/game src/main.c"
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
