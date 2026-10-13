# fe

Local middleware CLI for game dev and software dev — the glue between your engine/compiler and actually shipping.

```
fe init <name>    scaffold a new project
fe build          compile for all configured platforms
fe export         export all configured Godot presets
fe pack           pack sprites into a spritesheet
fe transcode      convert audio/video via ffmpeg
fe bmp            push builds to itch.io via butler
fe release        build → pack → push in one shot
fe validate       pre-flight config & path check
fe doctor         diagnose local toolchain and config issues
fe tui            interactive terminal UI (local only)
```

For Open 3D Engine (O3DE) workflows, fe also provides engine discovery and build helpers:

```sh
fe o3de doctor
fe o3de find engines
fe o3de find projects
fe o3de find gems
fe o3de find templates
fe o3de build --build-dir <configured-build-directory>
```

## Install

On Debian or Ubuntu, download the matching `.deb` package from the [latest release](https://github.com/flushwhy/fe/releases/latest) and install it with apt. Packages are published for 64-bit x86 and ARM systems:

```sh
# Debian / Ubuntu (x86-64)
curl -LO https://github.com/flushwhy/fe/releases/latest/download/fe_VERSION_amd64.deb
sudo apt install ./fe_VERSION_amd64.deb

# Debian / Ubuntu (ARM64): use the matching fe_VERSION_arm64.deb asset
```

Replace `VERSION` with the version shown in the release asset filename. These are downloadable release packages for now; a signed APT repository may be added later.

You can also download a standalone binary for your platform from the [releases page](https://github.com/flushwhy/fe/releases/latest):

```sh
# Linux x64
curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-linux-amd64 -o fe
chmod +x fe
sudo mv fe /usr/local/bin/

# Linux ARM64
curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-linux-arm64 -o fe
chmod +x fe
sudo mv fe /usr/local/bin/

# macOS arm64
curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-macos-arm64 -o fe
chmod +x fe
sudo mv fe /usr/local/bin/
```

On macOS, you can also install with Homebrew:

```sh
brew tap flushwhy/tap
brew install fe
```

Or build from source:

```sh
go install github.com/flushwhy/fe@latest
```

## Quick start

```sh
fe init go --name my-game
cd my-game
# edit .fe.yaml
fe build
fe export   # Godot projects: export all configured presets
fe release
```

Project scaffolds are downloaded from the [fe-templates repository](https://github.com/flushwhy/fe-templates) on first use and cached locally. An internet connection to GitHub is required the first time you use a template; cached templates can be reused offline. Browse templates with `fe template list`, or refresh one with `fe template update <template-id>`.

## Config — `.fe.yaml`

Config is optional. fe uses conventions when nothing is set.

```yaml
itchio:
  username: ""      # or env: ITCHIO_USERNAME
  game: ""          # or env: ITCHIO_GAME

butler:
  directory: "builds"       # root folder with platform subdirs
  userversion: ""           # or env: FE_USERVERSION

godot:
  executable: "godot"
  output_dir: "builds"

pack:
  input: "assets/sprites"   # default convention
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
```

### Config priority

```
env vars  >  .fe.yaml  >  conventions  >  defaults
```

| Env var | Overrides |
|--------------------|-------------------------|
| `ITCHIO_USERNAME` | `itchio.username` |
| `ITCHIO_GAME` | `itchio.game` |
| `FE_USERVERSION` | `butler.userversion` |
| `FE_BUILD_DIR` | `butler.directory` |

## CI usage

fe is designed to drop into any Git runner with zero extra config if you set the right secrets:

```yaml
# .github/workflows/release.yml
- name: Download fe
  run: |
    curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-linux-x64 -o fe
    chmod +x fe

- name: Release
  run: ./fe release --userversion ${{ github.ref_name }}
  env:
    ITCHIO_USERNAME: ${{ secrets.ITCHIO_USERNAME }}
    ITCHIO_GAME: ${{ secrets.ITCHIO_GAME }}
    BUTLER_API_KEY: ${{ secrets.BUTLER_API_KEY }}
```

## Godot exports

For Godot projects, configure export presets in Godot and commit `export_presets.cfg`. Then run `fe export` to export every preset, or `fe export --preset "Windows Desktop"` to export one exact preset. `godot` must be on your PATH (or set `godot.executable`), and the matching export templates/platform requirements must be installed. Preset output filenames/extensions are retained while `godot.output_dir` controls the output root.

Godot presets for Windows Desktop, Linux, and macOS are placed in `windows-x64/`, `linux-x64/`, and `macos-arm64/` respectively, ready for the existing Butler folder convention. Adjust `godot.output_dir` in `.fe.yaml` if you use a different build root. `fe release` automatically exports presets when run from a directory containing `project.godot`; non-Godot projects continue to use `build.targets`.

## Open 3D Engine (O3DE)

The `fe o3de` commands help locate local O3DE resources and work with an already configured CMake build directory. They are discovery/build helpers, not a replacement for installing or configuring O3DE.

### Diagnose your setup

```sh
fe o3de doctor
```

Checks the local O3DE setup and reports useful information about the detected environment.

### Find engines, projects, gems, and templates

```sh
fe o3de find engines
fe o3de find projects
fe o3de find gems
fe o3de find templates
```

Discovery uses registered O3DE manifest entries and bounded local filesystem search. Use `--path` to choose a search root, `--depth` to limit how far discovery descends, and `--registered-only` to restrict results to manifest-registered resources. For example:

```sh
fe o3de find projects --path ~/dev --depth 3
fe o3de find engines --registered-only
```

### Build an existing configured build directory

```sh
fe o3de build --build-dir /path/to/build
```

Pass the existing CMake build directory for your O3DE project. fe invokes the build and summarizes common failures; configure the project/build directory with the O3DE workflow first.

## Builds directory convention

`fe bmp` and `fe release` look for platform folders inside `builds/`:

```
builds/
├── windows-x64/    →  itch.io channel: windows-x64
├── linux-x64/      →  itch.io channel: linux-x64
└── macos-arm64/    →  itch.io channel: macos-arm64
```

Supported platforms: `windows`, `linux`, `macos`

Supported arches: `x64`, `x32`, `arm64`, `arm32`

## TUI

Run `fe tui` from an interactive terminal to open the Bubble Tea menu for common project, shipping, doctor, and configuration workflows. The TUI requires both keyboard input and terminal output, so it does not start in a non-interactive CI environment.

- While a command is running, press **Ctrl+C** to cancel that command and return to the TUI.
- When the TUI is idle, press **Ctrl+C** or **q** to exit.

The TUI calls the existing `fe` commands, so command behavior remains consistent with using the CLI directly.

## Dependencies

fe wraps existing tools — you need them installed separately:

| Command | Requires |
|---------------|------------------|
| `fe bmp` | [butler](https://itch.io/docs/butler/) |
| `fe transcode` | [ffmpeg](https://ffmpeg.org/) |
| `fe build` | whatever compiler your targets use |
| `fe o3de build` | an O3DE project with an existing configured CMake build directory |

## License

GPL-3.0
