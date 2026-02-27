#!/usr/bin/env bash
#
# build.sh — Cross-compile rocq-bootstrap.exe from Linux
#
# Prerequisites:
#   sudo apt install gcc-mingw-w64-x86-64 golang
#   go mod tidy  (run once)
#
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "==> Syncing embedded assets..."
cp -f ../manifest/latest.json embedded/manifest/latest.json
cp -f ../templates/test.v embedded/templates/test.v

echo "==> Building rocq-bootstrap.exe (Windows amd64)..."
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
GOOS=windows \
GOARCH=amd64 \
  go build \
    -ldflags="-H windowsgui -s -w" \
    -o rocq-bootstrap.exe \
    ./cmd/rocq-bootstrap/

echo "==> Done: $(ls -lh rocq-bootstrap.exe | awk '{print $5, $NF}')"
file rocq-bootstrap.exe
