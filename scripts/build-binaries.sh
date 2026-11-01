#!/usr/bin/env bash
set -euo pipefail

mkdir -p dist

ldflags="-s -w"
if [ -f internal/version/version.go ]; then
  version="${VERSION:-$(git describe --tags --exact-match 2>/dev/null || echo dev)}"
  version="${version#fe-v}"
  version="${version#v}"
  commit="$(git rev-parse HEAD)"
  ldflags="$ldflags -X github.com/flushwhy/fe/internal/version.Version=$version -X github.com/flushwhy/fe/internal/version.Commit=$commit"
fi

build() {
  local goos="$1" goarch="$2" asset="$3"
  echo "Building $asset ($goos/$goarch)"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags="$ldflags" -o "dist/$asset" .
}

if [ "$#" -eq 3 ]; then
  build "$@"
elif [ "$#" -eq 0 ]; then
  while read -r goos goarch asset; do
    build "$goos" "$goarch" "$asset"
  done <<'TARGETS'
linux amd64 fe-linux-amd64
linux arm64 fe-linux-arm64
windows amd64 fe-windows-amd64.exe
darwin amd64 fe-macos-amd64
darwin arm64 fe-macos-arm64
TARGETS
else
  echo "usage: $0 [GOOS GOARCH OUTPUT_NAME]" >&2
  exit 2
fi
