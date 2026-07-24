# fe

Local middleware CLI for game dev and software dev — the glue between your engine/compiler and actually shipping.

```
fe init <name>    scaffold a new project
fe build          compile for all configured platforms
fe pack           pack sprites into a spritesheet
fe transcode      convert audio/video via ffmpeg
fe bmp            push builds to itch.io via butler
fe release        build → pack → push in one shot
fe validate       pre-flight config & path check
fe tui            interactive terminal UI (local only)
```

## Install

Download the binary for your platform from the [releases page](https://github.com/flushwhy/fe/releases/latest):

```sh
# Linux x64
curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-linux-x64 -o fe
chmod +x fe
sudo mv fe /usr/local/bin/

# macOS arm64
curl -L https://github.com/flushwhy/fe/releases/latest/download/fe-macos-arm64 -o fe
chmod +x fe
sudo mv fe /usr/local/bin/
```

Or build from source:
```sh
go install github.com/flushwhy/fe@latest
```

## Quick start

```sh
fe init my-game
cd my-game
# edit .fe.yaml
fe build
fe release
```

## Config — `.fe.yaml`

Config is optional. fe uses conventions when nothing is set.

```yaml
itchio:
  username: ""      # or env: ITCHIO_USERNAME
  game: ""          # or env: ITCHIO_GAME

butler:
  directory: "builds"       # root folder with platform subdirs
  userversion: ""           # or env: FE_USERVERSION

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

| Env var            | Overrides               |
|--------------------|-------------------------|
| `ITCHIO_USERNAME`  | `itchio.username`       |
| `ITCHIO_GAME`      | `itchio.game`           |
| `FE_USERVERSION`   | `butler.userversion`    |
| `FE_BUILD_DIR`     | `butler.directory`      |

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

`fe tui` opens an interactive Bubble Tea interface. It auto-detects whether you're in a real terminal and refuses to run in CI.

## Dependencies

fe wraps existing tools — you need them installed separately:

| Command       | Requires         |
|---------------|------------------|
| `fe bmp`      | [butler](https://itch.io/docs/butler/) |
| `fe transcode`| [ffmpeg](https://ffmpeg.org/)          |
| `fe build`    | whatever compiler your targets use     |

## License

GPL-3.0
