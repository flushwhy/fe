<p align="center">
  <img src="https://raw.githubusercontent.com/MariaLetta/free-gophers-pack/master/PNG/128x128/gopher-rocket-fly.png" alt="A gopher riding a rocket" width="150">
</p>

<h1 align="center">Flush Exporter (fe)</h1>

<p align="center"><strong>A command-line tool to make the odd parts of game development easier.</strong></p>

<p align="center">
  <a href="https://dl.circleci.com/status-badge/redirect/gh/flushwhy/fe/tree/master"><img src="https://dl.circleci.com/status-badge/svg/gh/flushwhy/fe/tree/master" alt="CircleCI"></a>
  <a href="https://github.com/flushwhy/fe/releases"><img src="https://img.shields.io/github/v/release/flushwhy/fe" alt="GitHub release"></a>
  <a href="https://github.com/flushwhy/fe"><img src="https://img.shields.io/github/go-mod/go-version/flushwhy/fe" alt="Go version"></a>
  <a href="https://github.com/flushwhy/fe/blob/master/LICENSE"><img src="https://img.shields.io/github/license/flushwhy/fe" alt="License"></a>
  <a href="https://github.com/flushwhy/fe/stargazers"><img src="https://img.shields.io/github/stars/flushwhy/fe" alt="GitHub stars"></a>
  <a href="https://github.com/flushwhy/fe/issues"><img src="https://img.shields.io/github/issues/flushwhy/fe" alt="GitHub issues"></a>
</p>

---

## 🚀 Overview

**`fe`** is a CLI tool that streamlines tedious parts of game development—like asset processing and deployment—so you can focus on what matters: **making your game**.

---

## 🔧 Installation

### Prerequisites

You must have these tools installed and available in your system's `PATH`:

- [FFmpeg](https://ffmpeg.org): For audio/video transcoding.
- [Butler](https://itch.io/docs/butler): For publishing builds to itch.io.

To verify installation:

```bash
ffmpeg -version
butler -V
```

### Installing Dependencies

<details>
  <summary><strong>Windows</strong></summary>

```bash
# Install FFmpeg and Butler using winget
winget install ffmpeg

# Download and install Butler manually from itch.io
```

</details>

<details>
  <summary><strong>Linux (Debian/Ubuntu)</strong></summary>

```bash
# Install FFmpeg using apt
sudo apt update && sudo apt install ffmpeg

# Download and install Butler manually from itch.io
```

</details>

<details>
  <summary><strong>macOS</strong></summary>

```bash
# Install FFmpeg using Homebrew
brew install ffmpeg 

# Download and install Butler manually from itch.io
```

</details>

### Installing `fe`

Download the latest binary from the [Releases Page](https://github.com/flushwhy/fe/releases) and place the executable in a directory that's in your system's `PATH`.

---

## ⚙️ Configuration

Create a `.fe.yaml` file in your project root for simplified usage.

### Example `.fe.yaml`

```yaml
itchio:
  username: "your-itch-username"
  game: "your-itch-game-name"

pack:
  input: "./assets/raw_sprites"
  output: "./assets/spritesheet.png"

transcode:
  codec: "libvorbis"
  bitrate: "128k"
```

---

## 🎮 Usage

### Transcode Audio/Video

```bash
# Transcode a single file
fe transcode --inputFile sound.wav --outputFile sound.ogg --codec libvorbis

# Use config preset
fe transcode --inputFile assets/sounds/jump.wav --outputFile assets/sounds/jump.ogg
```

---

### Pack Sprites

```bash
# Pack PNGs into a spritesheet
fe pack --input ./assets/player_frames/ --output ./assets/player_sheet.png
```

---

### Push to Itch.io

```bash
# Push builds using config
fe bmp
```

---

### Init Game Project

```bash 
fe init MyNewGame
```
This build a file sturcture like:
```
MyNewGame/
├── .fe.yaml          # Default configuration for fe
├── assets/
│   ├── audio/        # For raw audio files (.wav, .mp3)
│   ├── fonts/        # For font files (.ttf, .otf)
│   └── sprites/      # For individual sprite images (.png)
├── builds/           # For your final, compiled game executables
└── src/              # For your game's source code

```
---

## 🗺️ Roadmap

- [x] Add audio/video transcoding  
- [x] Add bulk pusher for itch.io’s Butler  
- [x] Add texture packer (sprite sheet generator)  
- [ ] Add vector (SVG) to TTF font converter  
- [x] Add project scaffolding (`fe init`)  
- [ ] Add frameworks for things like Mesonbuild and raylib and so on to init. 
- [ ] Add file watcher for auto asset processing (`fe watch`)  
- [ ] More awesome stuff!
