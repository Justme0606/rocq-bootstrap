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

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck source=activate.sh
source "$SCRIPT_DIR/activate.sh"

exec "${SHELL:-bash}"
