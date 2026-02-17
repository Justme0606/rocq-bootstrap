#!/usr/bin/env bash
#
# rocq-bootstrap
# Reproducible and version-pinned Rocq environment bootstrapper.
#
# Copyright (c) 2026 Sylvain Borgogno
# Licensed under the MIT License.
#
# https://github.com/justme0606/rocq-bootstrap
#

LOG_DIR="${LOG_DIR:-$HOME/.rocq-setup/logs}"
LOG_FILE=""
VERBOSE=0
NON_INTERACTIVE=1
SKIP_VSCODE=0
WORKSPACE_DIR="${WORKSPACE_DIR:-$HOME/rocq-workspace}"
ROCQ_VERSION_ARG="latest"
WITH_ROCQIDE="no"
FORCE=0
TEST_ONLY=0
RECREATE_SWITCH=0
DOCTOR=0

log() { echo "[$(date +'%F %T')] $*" | tee -a "$LOG_FILE" >&2; }
die() { log "ERROR: $*"; exit 1; }

init_logging() {
  mkdir -p "$LOG_DIR"
  LOG_FILE="$LOG_DIR/rocq-setup-$(date +'%Y%m%d-%H%M%S').log"
  : > "$LOG_FILE"
  log "Log file: $LOG_FILE"
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --rocq-version) ROCQ_VERSION_ARG="${2:-}"; shift 2 ;;
      --workspace) WORKSPACE_DIR="${2:-}"; shift 2 ;;
      --with-rocqide) WITH_ROCQIDE="${2:-}"; shift 2 ;;
      --skip-vscode) SKIP_VSCODE=1; shift ;;
      --interactive) NON_INTERACTIVE=0; shift ;;
      --verbose) VERBOSE=1; shift ;;
      --force) FORCE=1; shift ;;
      --doctor) DOCTOR=1; shift ;;
      --test-only) TEST_ONLY=1; shift ;;
      --recreate-switch) RECREATE_SWITCH=1; shift ;;
      -h|--help)
        cat <<EOF
Usage: ./install.sh [options]
--rocq-version <x.y.z|latest>  (default: latest)
--workspace <path>             (default: ~/rocq-workspace)
--with-rocqide <yes|no>        (default: no)
--skip-vscode                  (default: false)
--interactive                  (default: non-interactive)
--doctor                      Run diagnostics only (no installation)
--verbose
--force
--test-only                   Run checks/tests only (no installation, no downloads)
--recreate-switch             Remove and recreate the opam switch if it already exists (Linux/opam)

EOF
        exit 0
        ;;
      *) die "Unknown argument: $1" ;;
    esac
  done
}

need_cmd() { command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"; }

download() {
  local url="$1" out="$2"
  need_cmd curl
  log "Downloading: $url"
  curl -fL --retry 3 --retry-delay 1 -o "$out" "$url"
}

sha256_check() {
  local file="$1" expected="$2"
  [[ -z "$expected" ]] && return 0
  need_cmd shasum
  local got
  got="$(shasum -a 256 "$file" | awk '{print $1}')"
  [[ "$got" == "$expected" ]] || die "SHA256 mismatch for $file (got=$got expected=$expected)"
}
