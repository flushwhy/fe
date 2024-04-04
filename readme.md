# <div style="text-align: center;">Flush Exporter</div>

---

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/flushwhy/fe)![GitHub license](https://img.shields.io/github/license/flushwhy/fe)[![GitHub release](https://img.shields.io/github/release/flushwhy/fe)](https://github.com/flushwhy/fe/releases) [![GitHub stars](https://img.shields.io/github/stars/flushwhy/fe)](https://github.com/flushwhy/fe/stargazers) [![GitHub forks](https://img.shields.io/github/forks/flushwhy/fe)](https://github.com/flushwhy/fe/network) [![GitHub issues](https://img.shields.io/github/issues/flushwhy/fe)](https://github.com/flushwhy/fe/issues) [![GitHub pull requests](https://img.shields.io/github/issues-pr/flushwhy/fe)](https://github.com/flushwhy/fe/pulls) [![GitHub contributors](https://img.shields.io/github/contributors/flushwhy/fe)](https://github.com/flushwhy/fe/graphs/contributors) [![GitHub last commit](https://img.shields.io/github/last-commit/flushwhy/fe)](https://github.com/flushwhy/fe/commits)

<div style="text-align: center;">

fe CLI is trying to make the odd parts of game development easier. Like it should be.
I started this project with the goal of making it easier to transcode audio/video and export your game to itchio.

The plan now to is expand on this project. To be a all in one CLI.

This is still a work in progress!!!

## Installation

You need FFmpeg, and itch.io's butler installed. Then you need to add the path to the ffmpeg and butler to your PATH. You can test by running the following into CMD, PowerShell, or Bash:

```bash
ffmpeg
```

```bash
butler
```

---

If your on Windows, You can can install FFmpeg and with the following commands:

open PowerShell and run the following:

```powershell
winget install ffmpeg
```

If you are on Linux:

```bash
sudo apt install ffmpeg
```

You will need to add fe to your PATH. If you don't put into the same directory as where you calling it form.

## Usage

```bash
fe transcode --inputFile <input> -outputFile <output>
```

or

```powershell
fe.exe transcode --inputFile <input> -outputFile <output>
```

## Plan

---

- [x] add audio conversion
- [x] add bulk pusher to itchio's bulter
  [] add sprite sheet compression
  [] add vector to TTF converter
  [] more stuff

</div>
