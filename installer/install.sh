#!/usr/bin/env bash
set -euo pipefail

# Installs the ccode wrapper script into ~/.local/bin/ccode. From a repository
# checkout it copies the adjacent wrapper; when piped to bash it downloads the
# wrapper from GitHub. It does not create config version files or release caches.

log_err() {
  printf 'install.sh: %s\n' "$*" >&2
}

die() {
  log_err "$*"
  exit 1
}

: "${HOME:?HOME is not set}"

WRAPPER_URL="https://raw.githubusercontent.com/cohesivestack/ccode/master/installer/bin/ccode"
SOURCE_WRAPPER=""
INSTALL_TMP_DIR=""

TARGET_DIR="${HOME}/.local/bin"
TARGET_WRAPPER="${TARGET_DIR}/ccode"

cleanup() {
  if [[ -n "$INSTALL_TMP_DIR" && -d "$INSTALL_TMP_DIR" ]]; then
    rm -rf -- "$INSTALL_TMP_DIR"
  fi
}

trap cleanup EXIT

resolve_source_wrapper() {
  local script_source="${BASH_SOURCE[0]:-}"

  if [[ -n "$script_source" && -f "$script_source" ]]; then
    local script_dir repo_root
    script_dir="$(cd -- "$(dirname -- "$script_source")" && pwd -P)"
    repo_root="$(cd -- "${script_dir}/.." && pwd -P)"
    SOURCE_WRAPPER="${repo_root}/installer/bin/ccode"
    [[ -f "$SOURCE_WRAPPER" ]] || die "wrapper source not found: ${SOURCE_WRAPPER}"
    return 0
  fi

  command -v curl >/dev/null 2>&1 || die "curl is required when installing from standard input"
  INSTALL_TMP_DIR="$(mktemp -d)"
  SOURCE_WRAPPER="${INSTALL_TMP_DIR}/ccode"
  curl -fsSL "$WRAPPER_URL" -o "$SOURCE_WRAPPER" || die "failed to download wrapper from ${WRAPPER_URL}"
  [[ -s "$SOURCE_WRAPPER" ]] || die "downloaded wrapper is empty: ${WRAPPER_URL}"
}

resolve_source_wrapper

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
