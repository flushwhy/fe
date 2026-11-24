# Fe

Fe is a local CLI that connects your game or software project to its existing tools. It helps scaffold projects, run configured builds, export Godot projects, prepare assets, and publish builds.

## Install

Download the latest release for your platform from the [GitHub releases page](https://github.com/flushwhy/fe/releases/latest).

### Linux

Debian and Ubuntu users can install the matching `.deb` package:

```sh
sudo apt install ./fe_VERSION_amd64.deb
```

Or install a standalone binary:

```sh
curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-linux-amd64 -o fe
chmod +x fe
sudo mv fe /usr/local/bin/
```

### macOS

Install with Homebrew:

```sh
brew tap flushwhy/tap
brew install fe
```

Or download the standalone macOS binary from the releases page.

### Windows

Install the stable release from Chocolatey:

```powershell
choco install fe
```

### Build from source

```sh
go install github.com/flushwhy/fe@latest
```

## Quick start

```sh
fe init go --name my-game
cd my-game
# Edit .fe.yaml for your project.
fe build
fe release
```

Project templates are downloaded from the [fe-templates repository](https://github.com/flushwhy/fe-templates) and cached locally.

## Project configuration

Fe reads an optional `.fe.yaml` file from the project directory. For example:

```yaml
itchio:
  username: ""
  game: ""

butler:
  directory: "builds"

godot:
  executable: "godot"
  output_dir: "builds"

pack:
  input: "assets/sprites"
  output: "assets/spritesheet.png"

build:
  targets:
    - name: linux-release
      platform: linux-x64
      cmd: "gcc -o builds/linux-x64/game src/main.c"
    - name: windows-release
      platform: windows-x64
      cmd: "x86_64-w64-mingw32-gcc -o builds/windows-x64/game.exe src/main.c"
```

Configuration values can also be supplied through environment variables. Environment variables take priority over `.fe.yaml`, followed by Fe's conventions and defaults.

## Common commands

| Command | Purpose |
|---|---|
| `fe init` | Create a project from a template. |
| `fe setup` | Set up supported project tools. |
| `fe build` | Run configured build targets. |
| `fe export` | Export configured Godot presets. |
| `fe pack` | Pack sprites into a spritesheet. |
| `fe transcode` | Convert audio and video with FFmpeg. |
| `fe bmp` | Upload build output to itch.io with Butler. |
| `fe release` | Run the configured release workflow. |
| `fe validate` | Validate project configuration. |
| `fe doctor` | Diagnose local tools and configuration. |

Run `fe --help` or `fe <command> --help` for the available options.

## Godot projects

Configure export presets in Godot and commit `export_presets.cfg`. Fe can export all configured presets or a selected preset:

```sh
fe export --list-presets
fe export
fe export --preset "Windows Desktop"
```

Godot must be available on your PATH, or configured with `godot.executable`. Matching export templates and platform requirements must also be installed.

## Build output

Fe expects platform-specific build output under `builds/` when publishing with Butler:

```text
builds/
├── windows-x64/
├── linux-x64/
└── macos-arm64/
```

## Dependencies

Fe uses existing tools rather than replacing them. Install the tools required by your project separately.

| Command | Requires |
|---|---|
| `fe bmp` | [Butler](https://itch.io/docs/butler/) |
| `fe transcode` | [FFmpeg](https://ffmpeg.org/) |
| `fe build` | The compiler or build system used by your targets |

## Releases

See the [release checklist](docs/releasing.md) for the stable release process.

## License

GPL-3.0
