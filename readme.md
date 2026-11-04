# fe

Local middleware CLI for game dev and software dev — the glue between your engine/compiler and actually shipping.

```
fe init <lang> --name <project>    scaffold a new project
fe build          compile for all configured platforms
fe export         export all configured Godot presets
fe pack           pack sprites into a spritesheet
fe transcode      convert audio/video via ffmpeg
fe bmp            push builds to itch.io via butler
fe release        build → pack → push in one shot
fe release --dry-run  preview the release plan without side effects
fe export --list-presets  inspect Godot export presets and output paths
fe manifest       write artifact inventory and SHA-256 checksums
fe changelog      generate Markdown release notes from Git history and tags
fe validate       pre-flight config & path check\nfe check [path]   run available static analyzers
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
# Multi-config generators and selected targets:
fe o3de build --build-dir build --config Profile --target MyProject
```

## Smart static checker

`fe check` detects supported source languages and runs their local static analyzers when available. Missing tools are reported with setup instructions; Fe does not install them automatically.

| Language | Checker | Enable it |
|---|---|---|
| Go | `go vet ./...` | Install Go from [go.dev/dl](https://go.dev/dl/) |
| C/C++ | `clang-tidy` | Install clang-tidy and configure CMake with `-DCMAKE_EXPORT_COMPILE_COMMANDS=ON` |
| Python | Ruff | `python -m pip install ruff` |
| JavaScript/TypeScript | ESLint | `npm install --save-dev eslint` |

Example: `fe check .`. Generated folders such as `node_modules`, `build`, and `dist` are excluded.

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

# Linux RISC-V 64-bit (native riscv64)
curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-linux-riscv64 -o fe
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

On Windows, install the stable release from Chocolatey:

```powershell
choco install fe
```

For the continuously updated development build, opt into prereleases of the same package:

```powershell
choco install fe --pre
```

The [Fe Rolling Release page](https://github.com/flushwhy/fe/releases/tag/rolling) is updated from successful builds on `master`. Rolling Chocolatey versions are prereleases based on the next patch after the latest stable release (for example, `1.2.4-rolling.gabc1234` after stable `1.2.3`, where `abc1234` is the commit SHA prefix). Chocolatey package versions are immutable, so each workflow run publishes a new prerelease version.

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
fe export --list-presets  # inspect Godot presets and output paths
fe export                 # export all configured presets
fe release
```

Project scaffolds are downloaded from the [fe-templates repository](https://github.com/flushwhy/fe-templates) on first use and cached locally. An internet connection to GitHub is required the first time you use a template; cached templates can be reused offline. Browse templates with `fe template list` to see their descriptions and categories; the last successful catalog is available offline. Refresh a cached template with `fe template update <template-id>`.

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
  # Video outputs default to H.264/AAC with game-friendly MP4 settings.
  # videoCodec: "libx264"
  # audioCodec: "aac"
  # preset: "medium"
  # crf: "20"
  # pixelFormat: "yuv420p"
  # audioBitrate: "192k"

build:
  targets:
    - name: linux-release
      platform: linux-x64
      cmd: "gcc -o builds/linux-x64/game src/main.c"
    - name: windows-release
      platform: windows-x64
      cmd: "x86_64-w64-mingw32-gcc -o builds/windows-x64/game.exe src/main.c"
    - name: macos-release
      platform: macos-arm64
      cmd: "gcc -o builds/macos-arm64/game src/main.c"
```

Select a named build target or every target for one platform:

```sh
fe build --list-targets
fe build --target linux-release
fe build --platform linux-x64
``

The optional `name` field lets projects define multiple configurations for the
same platform (for example, debug and release). Existing configurations that
only specify `platform` and `cmd` remain valid; `--platform` continues to
select all matching targets.

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

### Audio and video transcoding

`fe transcode` chooses defaults from the output extension. MP4/MOV video uses
H.264, AAC audio, CRF 20, the medium encoder preset, and `yuv420p` for broad
playback compatibility. MP4/MOV files are marked for fast-start streaming.
WebM uses VP9 and Opus. The source frame rate and resolution are preserved
unless explicitly changed.

```sh
# Game capture or trailer: compatible MP4 defaults
fe transcode --inputFile capture.mov --outputFile trailer.mp4

# Smaller web preview
fe transcode --inputFile capture.mp4 --outputFile preview.webm --crf 30

# Audio conversion
fe transcode --inputFile music.wav --outputFile music.ogg --codec libvorbis --bitrate 128k

# Override video settings
fe transcode --inputFile capture.mp4 --outputFile trailer.mp4 --crf 18 --preset slow --videoFrameRate 60
```

Use `--videoCodec`, `--audioCodec`, `--audioBitrate`, and
`--pixelFormat` to override the defaults. The older `--codec` and
`--bitrate` flags remain supported as media-type-specific overrides.

## CI usage

fe is designed to drop into any Git runner with zero extra config if you set the right secrets:

```yaml
# .github/workflows/release.yml
- name: Download fe
  run: |
    curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-linux-amd64 -o fe
    chmod +x fe

- name: Release
  run: ./fe release --userversion ${{ github.ref_name }}
  env:
    ITCHIO_USERNAME: ${{ secrets.ITCHIO_USERNAME }}
    ITCHIO_GAME: ${{ secrets.ITCHIO_GAME }}
    BUTLER_API_KEY: ${{ secrets.BUTLER_API_KEY }}
```

## Godot exports

For Godot projects, configure export presets in Godot and commit `export_presets.cfg`. Then run `fe export` to export every preset, `fe export --preset "Windows Desktop"` to export one exact preset, or `fe export --list-presets` to inspect preset names and expected output paths without exporting. `godot` must be on your PATH (or set `godot.executable`), and the matching export templates/platform requirements must be installed. Preset output filenames/extensions are retained while `godot.output_dir` controls the output root.

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

Pass the existing CMake build directory for your O3DE project. Use `--config Debug`, `--config Profile`, or `--config Release` with multi-config generators, and repeat `--target` to build selected targets. Fe invokes CMake and summarizes common failures; configure the project/build directory with the O3DE workflow first.

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

## Release checklist

Before publishing a stable version, use the [release checklist](docs/releasing.md) to verify CI, release assets, installation smoke tests, and the deployed documentation site.

## TUI

Run `fe tui` from an interactive terminal to open the Bubble Tea menu for common project, shipping, doctor, and configuration workflows. The TUI requires both keyboard input and terminal output, so it does not start in a non-interactive CI environment.

- While a command is running, press **Ctrl+C** to cancel that command and return to the TUI.
- When the TUI is idle, press **Ctrl+C** or **q** to exit.

The TUI calls the existing `fe` commands, so command behavior remains consistent with using the CLI directly.

## Watch mode

Run `fe watch` to monitor supported project files and trigger the configured development actions. Pass multiple directory names after one `--ignore-dir` flag to exclude extra directories, such as generated output or asset caches:

```sh
fe watch --ignore-dir generated cache
```

The directory names are collected until the next flag (starting with `--`), so you can still add other options afterward. Custom ignored directory names are matched case-insensitively and apply to nested directories, including matching directories created while watch mode is running.

## Release preview and artifact manifest

Run `fe release --dry-run` to inspect configured build/export targets, sprite packing, and itch.io publishing requirements without running build commands or uploading anything. The dry run reports missing credentials and unsupported Godot export platforms. It may still inspect local configuration and files, but does not build, pack, or publish.

`fe manifest` writes `builds/artifact-manifest.json` by default (or use `--dir`, `--output`, and `--version`). The JSON lists each artifact path, platform folder, file size, SHA-256 checksum, version, and commit when available. Symlinks are excluded and an empty artifact directory is treated as an error. `fe release` creates this manifest automatically before publishing. A failed sprite pack now stops the release instead of uploading potentially incomplete output.

`fe doctor` includes checks for Fe build target configuration and, in Godot projects, the configured Godot executable and export presets. Use `fe validate` for configuration validation and `fe doctor` for environment-oriented diagnostics.

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


## Archive packages

Use `fe package` to create a portable archive from the configured build directory:

```sh
fe package --format zip
fe package --dir builds/linux-x64 --format tar.gz --version 1.2.0
```

The command writes a versioned archive under `dist/` by default. ZIP and gzip-compressed TAR are supported; symlinks and non-regular files are excluded.

## Release orchestration

`fe release --dry-run` inspects the configured build/export targets, sprite packing, and itch.io publishing requirements without running commands or uploading anything. Optional release stages can generate notes and package artifacts:

```sh
fe release --changelog-output dist/RELEASE-NOTES.md --package-format zip
fe release --dry-run --changelog-output dist/RELEASE-NOTES.md --package-format tar.gz
```

Packaging and changelog generation are opt-in; omitting these flags preserves the existing release behavior.
