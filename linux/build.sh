#!/usr/bin/env bash
#
# build.sh — Build rocq-bootstrap for Linux
#
# Prerequisites:
#   sudo apt install golang libgl1-mesa-dev xorg-dev
#   go mod tidy  (run once)
#
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "==> Syncing embedded assets..."
cp -f ../manifest/latest.json embedded/manifest/latest.json
cp -f ../templates/test.v embedded/templates/test.v
cp -f ../templates/main.v embedded/templates/main.v
cp -f ../templates/_RocqProject embedded/templates/_RocqProject

echo "==> Building rocq-bootstrap (Linux amd64)..."
CGO_ENABLED=1 \
GOOS=linux \
GOARCH=amd64 \
  go build \
    -ldflags="-s -w" \
    -o rocq-bootstrap \
    ./cmd/rocq-bootstrap/

echo "==> Done: $(ls -lh rocq-bootstrap | awk '{print $5, $NF}')"
file rocq-bootstrap
