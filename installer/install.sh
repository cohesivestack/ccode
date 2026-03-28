#!/usr/bin/env bash
set -euo pipefail

# Installs the ccode wrapper script from this repository into ~/.local/bin/ccode.
# This installer only copies the wrapper and does not create config version files
# or release cache directories.

log_err() {
  printf 'install.sh: %s\n' "$*" >&2
}

die() {
  log_err "$*"
  exit 1
}

: "${HOME:?HOME is not set}"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd -P)"
SOURCE_WRAPPER="${REPO_ROOT}/installer/bin/ccode"

TARGET_DIR="${HOME}/.local/bin"
TARGET_WRAPPER="${TARGET_DIR}/ccode"

[[ -f "$SOURCE_WRAPPER" ]] || die "wrapper source not found: ${SOURCE_WRAPPER}"

mkdir -p "$TARGET_DIR"

if command -v install >/dev/null 2>&1; then
  install -m 0755 "$SOURCE_WRAPPER" "$TARGET_WRAPPER"
else
  cp "$SOURCE_WRAPPER" "$TARGET_WRAPPER"
  chmod 0755 "$TARGET_WRAPPER"
fi

printf 'Installed ccode wrapper to: %s\n' "$TARGET_WRAPPER"

if [[ ":${PATH:-}:" == *":${TARGET_DIR}:"* ]]; then
  printf 'PATH check: %s is already in PATH.\n' "$TARGET_DIR"
else
  printf 'PATH check: %s is not currently in PATH.\n' "$TARGET_DIR"
  printf 'Add it to your shell profile (for example ~/.zshrc or ~/.bashrc):\n'
  printf '  export PATH="%s:$PATH"\n' "$TARGET_DIR"
fi

printf 'Next step: ccode --help\n'
