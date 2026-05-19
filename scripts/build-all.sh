#!/usr/bin/env bash
# Cross-compile release binaries for common platforms.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

LDFLAGS='-s -w'
OUT="${ROOT}/dist"
mkdir -p "$OUT"

build() {
  local goos=$1 goarch=$2 name=$3
  echo "==> ${goos}/${goarch} -> ${name}"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -ldflags="$LDFLAGS" -o "$OUT/$name" .
}

build linux   amd64 lichess-puzzle-mixer-linux-amd64
build linux   arm64 lichess-puzzle-mixer-linux-arm64
build darwin  amd64 lichess-puzzle-mixer-macos-amd64
build darwin  arm64 lichess-puzzle-mixer-macos-arm64
build windows amd64 lichess-puzzle-mixer-windows-amd64.exe

echo ""
echo "Built:"
ls -lh "$OUT"
