# rocq-bootstrap

**Reproducible and version-pinned Rocq environment bootstrapper.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Status: In Development](https://img.shields.io/badge/status-in%20development-orange)

> **Note**: This project is currently under active development. Features and APIs may change.

## Overview

**rocq-bootstrap** is a cross-platform installation and environment bootstrap
tool designed to provide a reproducible and version-aligned development
environment for the Rocq Platform.

This project aims to simplify deployment for:

- Academic courses
- Research environments
- Workshops
- Student onboarding
- Controlled experimental setups

It enforces strict version alignment and deterministic installation,
closely following Rocq Platform release conventions.

---

## Objectives

The tool provides:

- A reproducible Rocq installation
- Strict version pinning across the Rocq stack
- Official repository prioritisation
- Automated workspace generation
- VSCode integration
- CLI validation of the installed toolchain

The design favours clarity, reproducibility and minimal manual
configuration.

---

## Supported Platforms

### Linux

On Linux systems, Rocq is installed using `opam`.

The installer:

- Creates a dedicated switch following Rocq Platform naming
  conventions

- Configures the official Rocq opam repository:

      https://rocq-prover.org/opam/released

- Installs a fully aligned Rocq stack:
  - `rocq-runtime=<version>`
  - `rocq-core=<version>`
  - `rocq-stdlib=<version>`
  - `rocq-prover=<version>`
  - `rocqide=<version>` (optional)
  - `vsrocq-language-server`

All core packages are strictly pinned to the requested Rocq version.

---

### macOS

On macOS systems, the installer:

- Resolves the appropriate signed Rocq Platform release asset
- Downloads the official signed installer
- Installs the application bundle
- Locates `vsrocqtop`
- Configures a ready-to-use workspace

Only signed release artifacts are accepted.

---

### Windows

On Windows, a standalone GUI installer (`rocq-bootstrap.exe`) handles
the entire setup:

- Downloads the official signed Rocq Platform InnoSetup installer
- Verifies SHA256 checksum
- Runs the InnoSetup installer silently
- Locates `vsrocqtop` in the installation directory
- Installs the VSRocq extension in VSCode
- Creates a ready-to-use workspace in `%USERPROFILE%\rocq-workspace`
- Configures VSCode settings and opens the workspace

The installer is a Go application with an embedded GUI (Fyne) that
displays real-time progress. It embeds the manifest and workspace
templates at build time.

**Default installation directory:**

    C:\Rocq-platform~<rocq_major.minor>~<platform_year.month>

Example: `C:\Rocq-platform~9.0~2025.08`

The installer automatically detects existing installations (via
filesystem checks, Windows registry, and PATH lookup) and skips
the download/install steps when an existing installation is found.

---

## Reproducibility Model

The installation process is driven by a manifest file:

    manifest/latest.json

This file specifies:

- Platform release identifier
- Rocq version
- Optional snapshot identifier
- macOS and Windows release assets
- SHA256 checksums

The manifest guarantees:

- Version consistency
- Controlled dependency resolution
- Explicit release targeting

---

## Installation Requirements

### Linux

- `opam ≥ 2.1`
- `jq`
- `curl`
- VSCode (optional but recommended)

### macOS

- `curl`
- `jq`
- VSCode (optional but recommended)

### Windows

No prerequisites for end users — just run `rocq-bootstrap.exe`.

For building from source (cross-compilation from Linux):

- `go >= 1.22`
- `gcc-mingw-w64-x86-64`

---

## Usage

### Standard installation

```bash
./install.sh
```

---

### Specify a workspace directory

```bash
./install.sh --workspace /path/to/workspace
```

---

### Recreate switch (Linux only)

```bash
./install.sh --recreate-switch
```

This removes and recreates the opam switch to ensure a clean
environment.

---

### Windows

Simply run the GUI installer:

    rocq-bootstrap.exe

To build from source (cross-compile from Linux):

```bash
cd windows
make all       # production build (no console window)
make debug     # debug build (shows console for error output)
make clean     # remove build artifacts
```

---

### Test-only mode

```bash
./install.sh --test-only
```

This mode:

- Resolves the manifest
- Prepares the workspace
- Runs validation tests
- Does not install Rocq

---

## Workspace Structure

The installer generates:

    <workspace>/
     ├── test.v
     ├── _RocqProject
     └── .vscode/
         └── settings.json

The workspace is configured to:

- Use the installed `vsrocqtop`
- Open directly in VSCode
- Compile a minimal validation file

---

## Validation Procedure

Upon completion, the installer performs a CLI validation:

    rocq compile test.v

Installation is considered successful only if compilation succeeds.

---

## Switch Naming Convention (Linux)

Switches follow the Rocq Platform naming scheme:

    CP.<platform_release>~<rocq_major.minor>

Examples:

    CP.2025.08.1~9.0
    CP.2025.08.1~9.0~2025.08

This ensures traceability and compatibility with official release
patterns.

---

## Intended Audience

This tool is primarily intended for:

- Research groups
- Teaching staff
- Graduate and undergraduate courses
- Controlled Rocq environments requiring reproducibility

---

## License

Copyright (c) 2026 Sylvain Borgogno

Licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Repository

https://github.com/justme0606/rocq-bootstrap
