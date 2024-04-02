# Flush Exporter

---
fe is cli to help with with gamedev.


## Installation

You need FFmpeg installed.

if on Windows:

open PowerShell and run the following:

```powershell
winget install ffmpeg
```

if on Linux:

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
- [] add bulk pusher to itchio's bulter
- [] add sprite sheet compression
- [] add vector to TTF converter
- [] more stuff
