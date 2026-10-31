#!/usr/bin/env bash
set -euo pipefail

mkdir -p dist

while read -r goos goarch asset; do
  echo "Building $asset ($goos/$goarch)"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags="-s -w" -o "dist/$asset" .
done <<'TARGETS'
linux amd64 fe-linux-amd64
linux arm64 fe-linux-arm64
windows amd64 fe-windows-amd64.exe
darwin amd64 fe-macos-amd64
darwin arm64 fe-macos-arm64
TARGETS
