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

install_rocq_linux_opam() {
  need_cmd opam

  local opam_ver
  opam_ver="$(opam --version | tr -d '\r')"
  log "opam version: $opam_ver"
  [[ "$opam_ver" == 2.* ]] || die "opam >= 2.1.0 required (found $opam_ver)"

  if [[ ! -d "$HOME/.opam" ]]; then
    log "Initializing opam..."
    opam init -y --bare
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