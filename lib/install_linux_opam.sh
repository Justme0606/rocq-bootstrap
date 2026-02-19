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

set -euo pipefail

REPO_NAME="rocq-released"
REPO_URL="https://rocq-prover.org/opam/released"

ensure_opam_deps() {
  # command -> package name mapping (command:package)
  local deps="unzip:unzip bwrap:bubblewrap make:make cc:gcc"
  local missing_pkgs=()

  for entry in $deps; do
    local cmd="${entry%%:*}"
    local pkg="${entry##*:}"
    command -v "$cmd" >/dev/null 2>&1 || missing_pkgs+=("$pkg")
  done

  [[ ${#missing_pkgs[@]} -eq 0 ]] && return 0

  log "Installing opam dependencies: ${missing_pkgs[*]}"

  local SUDO=""
  if [[ "$(id -u)" -ne 0 ]]; then
    command -v sudo >/dev/null 2>&1 || die "Cannot install opam dependencies (${missing_pkgs[*]}): not root and sudo not available. Please install them manually."
    SUDO="sudo"
  fi

  if command -v apt-get >/dev/null 2>&1; then
    $SUDO apt-get update -qq >&2 && $SUDO apt-get install -y -qq "${missing_pkgs[@]}" >&2
  elif command -v dnf >/dev/null 2>&1; then
    $SUDO dnf install -y "${missing_pkgs[@]}" >&2
  elif command -v yum >/dev/null 2>&1; then
    $SUDO yum install -y "${missing_pkgs[@]}" >&2
  elif command -v pacman >/dev/null 2>&1; then
    $SUDO pacman -S --noconfirm "${missing_pkgs[@]}" >&2
  elif command -v zypper >/dev/null 2>&1; then
    $SUDO zypper install -y "${missing_pkgs[@]}" >&2
  else
    die "Cannot install opam dependencies (${missing_pkgs[*]}): no supported package manager found. Please install them manually."
  fi
}

ensure_opam() {
  if command -v opam >/dev/null 2>&1; then
    return 0
  fi

  log "opam not found — installing via official installer..."
  need_cmd curl

  ensure_opam_deps

  local tmp
  tmp="$(mktemp)"
  curl -fL --retry 3 --retry-delay 1 -o "$tmp" https://opam.ocaml.org/install.sh
  chmod +x "$tmp"
  sh "$tmp" --no-backup >&2
  rm -f "$tmp"

  # The installer places opam in /usr/local/bin or ~/.opam/bin; verify it worked
  if ! command -v opam >/dev/null 2>&1; then
    # Try common install location
    export PATH="/usr/local/bin:$HOME/.opam/bin:$PATH"
    command -v opam >/dev/null 2>&1 || die "opam installation failed — please install opam manually: https://opam.ocaml.org/doc/Install.html"
  fi

  log "opam installed successfully: $(command -v opam)"
}

install_rocq_linux_opam() {
  ensure_opam

  local opam_ver
  opam_ver="$(opam --version | tr -d '\r')"
  log "opam version: $opam_ver"
  [[ "$opam_ver" == 2.* ]] || die "opam >= 2.1.0 required (found $opam_ver)"

  if [[ ! -d "$HOME/.opam" ]]; then
    log "Initializing opam..."
    opam init -y --bare --disable-sandboxing
  fi

  # Switch name style Rocq Platform
  local rocq_mm="${ROCQ_VERSION%.*}"  # 9.0 from 9.0.0
  local switch="CP.${PLATFORM_RELEASE}~${rocq_mm}"
  if [[ -n "${OPAM_SNAPSHOT:-}" ]]; then
    switch="${switch}~${OPAM_SNAPSHOT}"
  fi
  
   OPAM_SWITCH_NAME="$switch"

  # Recreate switch if requested
  if opam switch list --short | grep -qx "$switch"; then
    if [[ "${RECREATE_SWITCH:-0}" -eq 1 ]]; then
      log "Recreating opam switch (requested): $switch"
      opam switch remove "$switch" -y
    else
      log "Opam switch already exists: $switch"
      log "Tip: re-run with --recreate-switch to start from a clean switch"
    fi
  fi

  # Create switch if missing
  if ! opam switch list --short | grep -qx "$switch"; then
    log "Creating opam switch: $switch (ocaml 4.14.2)"
    opam switch create "$switch" ocaml-base-compiler.4.14.2 -y
  fi

  # Ensure rocq-released repo is present and correctly configured IN THIS SWITCH
  log "Ensuring opam repo $REPO_NAME -> $REPO_URL (switch=$switch)"
  if ! opam repo add --switch="$switch" "$REPO_NAME" "$REPO_URL" -y; then
    log "$REPO_NAME already exists; forcing set-url to $REPO_URL"
    opam repo set-url --switch="$switch" "$REPO_NAME" "$REPO_URL" -y
  fi

  # Force repo selection for THIS switch (avoid coq-local surprises)
  # Keep only what we want for reproducibility.
  opam repo set-repos --switch="$switch" "$REPO_NAME" default archive

  # Make rocq-released top priority (opam 2.5 syntax)
  opam repo priority --switch="$switch" "$REPO_NAME" 1
  opam update --switch="$switch"

  log "Installing Rocq stack pinned to $ROCQ_VERSION in switch $switch"
  opam install --switch="$switch" -y \
    "rocq-runtime=$ROCQ_VERSION" \
    "rocq-core=$ROCQ_VERSION" \
    "rocq-stdlib=$ROCQ_VERSION" \
    "rocq-prover=$ROCQ_VERSION"

  # Install vsrocq language server (provides vsrocqtop) unless VSCode is skipped
  if [[ "${SKIP_VSCODE:-0}" -eq 1 ]]; then
    log "SKIP_VSCODE=1: not installing vsrocq-language-server (vsrocqtop not required)"
  else
    log "Installing vsrocq-language-server (provides vsrocqtop)"
    opam install --switch="$switch" -y "vsrocq-language-server=2.3.4" || \
      opam install --switch="$switch" -y vsrocq-language-server
  fi

  if [[ "${WITH_ROCQIDE:-no}" == "yes" ]]; then
    log "Installing rocqide"
    opam install --switch="$switch" -y "rocqide=$ROCQ_VERSION"
  else
    log "Skipping rocqide (WITH_ROCQIDE=${WITH_ROCQIDE:-no})"
  fi

  local bin
  bin="$(opam var --switch="$switch" bin)"

  VSROCQTOP_PATH="$bin/vsrocqtop"
  ROCQ_PATH="$bin/rocq"

  # vsrocqtop is only required if we configure VSCode
  if [[ ! -x "$VSROCQTOP_PATH" ]]; then
    if [[ "${SKIP_VSCODE:-0}" -eq 1 ]]; then
      log "SKIP_VSCODE=1: vsrocqtop not found (ok)"
      VSROCQTOP_PATH=""
    else
      die "vsrocqtop not found in switch bin: $VSROCQTOP_PATH (install vsrocq-language-server)"
    fi
  fi

  [[ -x "$ROCQ_PATH" ]] || die "rocq not found in switch bin: $ROCQ_PATH"

  log "vsrocqtop: ${VSROCQTOP_PATH:-<none>}"
  log "rocq: $ROCQ_PATH"

  # Verify version matches requested (major.minor)
  local got want_mm
  got="$("$ROCQ_PATH" --print-version 2>/dev/null || true)"
  if [[ -z "$got" ]]; then
    got="$("$ROCQ_PATH" --version 2>&1 | head -n 1 || true)"
  fi
  log "rocq version after install: $got"

  want_mm="${ROCQ_VERSION%.*}" # 9.0
  echo "$got" | grep -q "$want_mm" || die "Installed rocq does not match requested Rocq $ROCQ_VERSION (got: $got)"

}