# rocq-bootstrap

**Reproducible and version-pinned Rocq environment bootstrapper.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Status: In Development](https://img.shields.io/badge/status-in%20development-orange)

> **Note**: This project is currently under active development. Features and APIs may change.

## Overview

**rocq-bootstrap** is a cross-platform installation and environment bootstrap
tool designed to provide a reproducible and version-aligned development
environment for the Rocq Platform (formerly Coq).

It supports both **Rocq** (version 9+) and **Coq** (version < 9) releases,
automatically selecting the correct packages, extensions, and binaries
depending on the chosen release.

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
- VSCode integration (VSRocq for Rocq 9+, VSCoq for Coq < 9)
- CLI validation of the installed toolchain

The design favours clarity, reproducibility and minimal manual
configuration.

---

## Supported Platforms

### Linux

On Linux systems, Rocq is installed using `opam`. Two installation
methods are available:

**GUI installer** (`rocq-bootstrap`): A standalone graphical installer
(Fyne) that handles the entire setup:

- Checks for opam and initialises it if needed
- Creates a dedicated opam switch following Rocq Platform naming
  conventions
- Configures the official Rocq opam repository
- Installs all Rocq/Coq packages with version pinning
- Creates a workspace with activation scripts (`activate.sh`,
  `activate-shell.sh`)
- Installs the appropriate VSCode extension (VSRocq for Rocq 9+,
  VSCoq for Coq < 9) and opens the workspace

The GUI displays real-time progress and includes a Doctor diagnostic
button.

**Shell installer** (`install.sh`): The original script-based installer
for headless or automated setups.

Both methods:

- Create a dedicated switch following Rocq Platform naming
  conventions

- Configure the official Rocq opam repository:

      https://rocq-prover.org/opam/released

- Install a fully aligned package stack depending on the version:

  **Rocq 9+:**
  - `rocq-runtime=<version>`
  - `rocq-core=<version>`
  - `rocq-stdlib=<version>`
  - `rocq-prover=<version>`
  - `rocqide=<version>` (optional)
  - `vsrocq-language-server`

  **Coq < 9:**
  - `coq=<version>`
  - `coqide=<version>` (optional)
  - `vscoq-language-server`

All core packages are strictly pinned to the requested version.

---

### macOS

On macOS systems, the installer:

- Resolves the appropriate signed Rocq Platform release asset
- Downloads the official signed installer
- Installs the application bundle
- Locates the language server binary (`vsrocqtop` for Rocq 9+,
  `vscoqtop` for Coq < 9)
- Configures a ready-to-use workspace

Only signed release artifacts are accepted.

---

### Windows

On Windows, a standalone GUI installer (`rocq-bootstrap.exe`) handles
the entire setup:

- Downloads the official signed Rocq Platform InnoSetup installer
- Verifies SHA256 checksum
- Runs the InnoSetup installer silently
- Locates the language server binary (`vsrocqtop` or `vscoqtop`)
- Installs the appropriate VSCode extension (VSRocq or VSCoq)
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

For the shell installer (`install.sh`):

- `opam ≥ 2.1`
- `jq`
- `curl`
- VSCode (optional but recommended)

For the GUI installer (`rocq-bootstrap`): no prerequisites for end
users — just run the binary. opam will be detected automatically.

For building the GUI from source:

- `go >= 1.22`
- Fyne system dependencies: `libgl-dev libxxf86vm-dev libxi-dev
  libxcursor-dev libxrandr-dev libxinerama-dev`

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

### Linux (GUI)

Download `rocq-bootstrap-linux-x86_64` from the
[Releases](https://github.com/justme0606/rocq-bootstrap/releases)
page, then:

```bash
chmod +x rocq-bootstrap-linux-x86_64

# Run directly (no installation needed)
./rocq-bootstrap-linux-x86_64

# Or install as a desktop application (icon in app menu)
./rocq-bootstrap-linux-x86_64 --install

# Uninstall
rocq-bootstrap --uninstall
```

`--install` copies the binary to `~/.local/bin/`, the Rocq icon to
`~/.local/share/icons/`, and creates a `.desktop` entry so the
application appears in your desktop environment's application menu.

To build from source:

```bash
cd linux
make all       # production build
make install   # install to ~/.local (binary, icon, .desktop)
make uninstall # remove from ~/.local
make clean     # remove build artifacts
```

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

- Use the installed language server (`vsrocqtop` or `vscoqtop`)
- Open directly in VSCode with the correct extension settings
  (`vsrocq.path` or `vscoq.path`)
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
