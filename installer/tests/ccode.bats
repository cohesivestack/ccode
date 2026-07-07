#!/usr/bin/env bats

WRAPPER=""
TEST_ROOT=""
TEST_PLATFORM=""
TEST_BINARY_NAME=""
TEST_ARCHIVE_EXT=""
MOCK_BIN=""

detect_platform() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "$os" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) os="windows" ;;
    *) echo "unsupported test OS: $os" >&2; return 1 ;;
  esac

  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) echo "unsupported test arch: $arch" >&2; return 1 ;;
  esac

  TEST_PLATFORM="${os}_${arch}"
  if [[ "$os" == "windows" ]]; then
    TEST_BINARY_NAME="ccode.exe"
    TEST_ARCHIVE_EXT="zip"
  else
    TEST_BINARY_NAME="ccode"
    TEST_ARCHIVE_EXT="tar.gz"
  fi
}

make_cached_binary() {
  local version="$1"
  local dir="${XDG_CACHE_HOME}/ccode/releases/${version}/${TEST_PLATFORM}"
  local bin="${dir}/${TEST_BINARY_NAME}"
  mkdir -p "$dir"
  cat >"$bin" <<EOF
#!/usr/bin/env bash
printf 'BIN=%s\n' '${version}'
for arg in "\$@"; do
  printf 'ARG=%s\n' "\$arg"
done
EOF
  chmod +x "$bin"
}

run_wrapper() {
  local dir="$1"
  shift
  (
    cd "$dir"
    "$WRAPPER" "$@"
  )
}

make_mock_curl() {
  cat >"${MOCK_BIN}/curl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

out_file=""
url=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--output) out_file="$2"; shift 2 ;;
    -H|--header|--retry|--retry-delay) shift 2 ;;
    -f|-s|-S|-L|-k|--fail|--silent|--show-error|--location) shift ;;
    -fsSL|-fL|-sSL|-fsS|-sS|-SL) shift ;;
    *)
      url="$1"
      shift
      ;;
  esac
done

if [[ -n "${MOCK_CURL_LOG:-}" && -n "$url" ]]; then
  printf '%s\n' "$url" >>"$MOCK_CURL_LOG"
fi

case "$url" in
  *"/releases/latest")
    cat <<JSON
{
  "tag_name": "${MOCK_LATEST_TAG:-v9.9.9}",
  "draft": false,
  "prerelease": false
}
JSON
    ;;
  *"/releases/tags/"*)
    tag="${url##*/releases/tags/}"
    version="${tag#v}"
    asset="ccode_${version}_${MOCK_PLATFORM}.${MOCK_EXT}"
    download_url="https://example.test/${asset}"
    cat <<JSON
{
  "assets": [
    {
      "name": "${asset}",
      "browser_download_url": "${download_url}"
    }
  ]
}
JSON
    ;;
  https://example.test/*)
    [[ -n "$out_file" ]] || {
      echo "mock curl expected -o for download URL" >&2
      exit 1
    }
    printf 'mock-archive' >"$out_file"
    ;;
  *)
    echo "mock curl unexpected URL: ${url}" >&2
    exit 1
    ;;
esac
EOF
  chmod +x "${MOCK_BIN}/curl"
}

make_mock_tar() {
  cat >"${MOCK_BIN}/tar" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

extract_dir=""
archive=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -C) extract_dir="$2"; shift 2 ;;
    -xzf|-xf|-x|-z|-f) shift ;;
    --*) shift ;;
    *)
      if [[ -z "$archive" ]]; then
        archive="$1"
      fi
      shift
      ;;
  esac
done

[[ -n "$extract_dir" ]] || {
  echo "mock tar missing -C extract dir" >&2
  exit 1
}

mkdir -p "$extract_dir"
base="$(basename "$archive")"
version="${base#ccode_}"
version="${version%%_*}"

cat >"${extract_dir}/ccode" <<SCRIPT
#!/usr/bin/env bash
printf 'BIN=v%s\n' '${version}'
for arg in "\$@"; do
  printf 'ARG=%s\n' "\$arg"
done
SCRIPT
chmod +x "${extract_dir}/ccode"
EOF
  chmod +x "${MOCK_BIN}/tar"
}

make_mock_git() {
  cat >"${MOCK_BIN}/git" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${FAKE_GIT_ROOT:-}" != "" && "${1:-}" == "-C" && "${3:-}" == "rev-parse" && "${4:-}" == "--show-toplevel" ]]; then
  printf '%s\n' "$FAKE_GIT_ROOT"
  exit 0
fi
exit 1
EOF
  chmod +x "${MOCK_BIN}/git"
}

setup() {
  WRAPPER="${BATS_TEST_DIRNAME}/../bin/ccode"
  TEST_ROOT="$(mktemp -d "${BATS_TEST_TMPDIR:-/tmp}/ccode-bats.XXXXXX")"
  export HOME="${TEST_ROOT}/home"
  export XDG_CONFIG_HOME="${HOME}/.config"
  export XDG_CACHE_HOME="${HOME}/.cache"
  export WORKSPACE="${TEST_ROOT}/workspace"
  mkdir -p "$HOME" "$XDG_CONFIG_HOME" "$XDG_CACHE_HOME" "$WORKSPACE"

  detect_platform

  MOCK_BIN="${TEST_ROOT}/mock-bin"
  mkdir -p "$MOCK_BIN"
  export PATH="${MOCK_BIN}:${PATH}"

  export MOCK_CURL_LOG="${TEST_ROOT}/mock-curl.log"
  export MOCK_LATEST_TAG="v9.9.9"
  export MOCK_PLATFORM="$TEST_PLATFORM"
  export MOCK_EXT="$TEST_ARCHIVE_EXT"
  : >"$MOCK_CURL_LOG"
  unset FAKE_GIT_ROOT

  make_mock_curl
  make_mock_tar
  make_mock_git
}

teardown() {
  rm -rf "$TEST_ROOT"
}

@test "pin v1.2.3 writes global version file" {
  run run_wrapper "$WORKSPACE" pin v1.2.3
  [ "$status" -eq 0 ]
  [ "$output" = "v1.2.3" ]
  [ "$(cat "$XDG_CONFIG_HOME/ccode/version")" = "v1.2.3" ]
  [ ! -e "$WORKSPACE/.ccode/version" ]
}

@test "pin 1.2.3 normalizes to v1.2.3" {
  run run_wrapper "$WORKSPACE" pin 1.2.3
  [ "$status" -eq 0 ]
  [ "$output" = "v1.2.3" ]
  [ "$(cat "$XDG_CONFIG_HOME/ccode/version")" = "v1.2.3" ]
}

@test "pin rejects --global because pin is always global" {
  run run_wrapper "$WORKSPACE" pin v1.2.3 --global
  [ "$status" -ne 0 ]
  [[ "$output" == *"unknown argument for pin: --global"* ]]
}

@test "pin with no version resolves latest stable and writes global version file" {
  export MOCK_LATEST_TAG="v1.10.0"

  run run_wrapper "$WORKSPACE" pin
  [ "$status" -eq 0 ]
  [ "$output" = "v1.10.0" ]
  [ "$(cat "$XDG_CONFIG_HOME/ccode/version")" = "v1.10.0" ]
  grep -q "/releases/latest" "$MOCK_CURL_LOG"
}

@test "pin latest resolves latest stable from mocked GitHub and writes concrete tag" {
  export MOCK_LATEST_TAG="v2.4.6"

  run run_wrapper "$WORKSPACE" pin latest
  [ "$status" -eq 0 ]
  [ "$output" = "v2.4.6" ]
  [ "$(cat "$XDG_CONFIG_HOME/ccode/version")" = "v2.4.6" ]
  grep -q "/releases/latest" "$MOCK_CURL_LOG"
}

@test "pin rejects unknown flag" {
  run run_wrapper "$WORKSPACE" pin --nope
  [ "$status" -ne 0 ]
  [[ "$output" == *"unknown argument for pin"* ]]
}

@test "pin rejects too many positional args" {
  run run_wrapper "$WORKSPACE" pin v1.2.3 v1.2.4
  [ "$status" -ne 0 ]
  [[ "$output" == *"too many positional arguments for pin"* ]]
}

@test "pin rejects invalid explicit version" {
  run run_wrapper "$WORKSPACE" pin nope
  [ "$status" -ne 0 ]
  [[ "$output" == *"invalid version for pin"* ]]
}

@test "normal execution uses --version over ccode.yaml and global pin" {
  make_cached_binary v3.0.0
  make_cached_binary v1.1.1
  printf 'ccode_path: ccode\nversion: v1.1.1\n' >"$WORKSPACE/ccode.yaml"
  mkdir -p "$XDG_CONFIG_HOME/ccode"
  printf 'v1.1.1\n' >"$XDG_CONFIG_HOME/ccode/version"

  run run_wrapper "$WORKSPACE" --version v3.0.0 run processA
  [ "$status" -eq 0 ]
  [[ "$output" == *"BIN=v3.0.0"* ]]
  [[ "$output" == *"ARG=run"* ]]
  [[ "$output" == *"ARG=processA"* ]]
  [[ "$output" != *"ARG=--version"* ]]
}

@test "normal execution uses nearest ccode.yaml version over global pin" {
  make_cached_binary v1.2.0
  make_cached_binary v4.0.0
  mkdir -p "$XDG_CONFIG_HOME/ccode"
  printf 'v4.0.0\n' >"$XDG_CONFIG_HOME/ccode/version"

  local root="${WORKSPACE}/root"
  local nested="${root}/a/b"
  mkdir -p "$nested"
  printf 'ccode_path: ccode\nversion: v1.2.0\n' >"${root}/ccode.yaml"

  run run_wrapper "$nested" run task
  [ "$status" -eq 0 ]
  [[ "$output" == *"BIN=v1.2.0"* ]]
}

@test "normal execution uses global pin when ccode.yaml is absent" {
  make_cached_binary v5.0.0
  mkdir -p "$XDG_CONFIG_HOME/ccode"
  printf 'v5.0.0\n' >"$XDG_CONFIG_HOME/ccode/version"

  run run_wrapper "$WORKSPACE" run task
  [ "$status" -eq 0 ]
  [[ "$output" == *"BIN=v5.0.0"* ]]
}

@test "normal execution resolves latest stable when no version source exists" {
  export MOCK_LATEST_TAG="v7.7.7"

  run run_wrapper "$WORKSPACE" run task --flag=value
  [ "$status" -eq 0 ]
  [[ "$output" == *"BIN=v7.7.7"* ]]
  [[ "$output" == *"ARG=run"* ]]
  [[ "$output" == *"ARG=task"* ]]
  [[ "$output" == *"ARG=--flag=value"* ]]
  grep -q "/releases/latest" "$MOCK_CURL_LOG"
  grep -q "/releases/tags/v7.7.7" "$MOCK_CURL_LOG"
}

@test "normal execution does not create local or global version files" {
  export MOCK_LATEST_TAG="v1.0.0"
  run run_wrapper "$WORKSPACE" run task
  [ "$status" -eq 0 ]
  [ ! -f "$WORKSPACE/.ccode/version" ]
  [ ! -f "$XDG_CONFIG_HOME/ccode/version" ]
}

@test "arguments are forwarded unchanged to the resolved binary" {
  make_cached_binary v2.0.0
  printf 'ccode_path: ccode\nversion: v2.0.0\n' >"$WORKSPACE/ccode.yaml"

  run run_wrapper "$WORKSPACE" --alpha "two words" --beta=3 "x=y"
  [ "$status" -eq 0 ]
  [[ "$output" == *"BIN=v2.0.0"* ]]
  [[ "$output" == *"ARG=--alpha"* ]]
  [[ "$output" == *"ARG=two words"* ]]
  [[ "$output" == *"ARG=--beta=3"* ]]
  [[ "$output" == *"ARG=x=y"* ]]
}

@test "init without --version forwards existing ccode.yaml version" {
  export MOCK_LATEST_TAG="v8.0.0"
  make_cached_binary v1.1.1
  printf 'ccode_path: ccode\nversion: v1.1.1\n' >"$WORKSPACE/ccode.yaml"
  mkdir -p "$XDG_CONFIG_HOME/ccode"
  printf 'v8.0.0\n' >"$XDG_CONFIG_HOME/ccode/version"

  run run_wrapper "$WORKSPACE" init project
  [ "$status" -eq 0 ]
  [[ "$output" == *"BIN=v1.1.1"* ]]
  [[ "$output" == *"ARG=init"* ]]
  [[ "$output" == *"ARG=project"* ]]
  [[ "$output" == *"ARG=--version"* ]]
  [[ "$output" == *"ARG=v1.1.1"* ]]
}

@test "init without ccode.yaml version forwards latest stable" {
  export MOCK_LATEST_TAG="v8.0.0"
  printf 'ccode_path: ccode\n' >"$WORKSPACE/ccode.yaml"
  mkdir -p "$XDG_CONFIG_HOME/ccode"
  printf 'v1.1.1\n' >"$XDG_CONFIG_HOME/ccode/version"

  run run_wrapper "$WORKSPACE" init project
  [ "$status" -eq 0 ]
  [[ "$output" == *"BIN=v8.0.0"* ]]
  [[ "$output" == *"ARG=init"* ]]
  [[ "$output" == *"ARG=project"* ]]
  [[ "$output" == *"ARG=--version"* ]]
  [[ "$output" == *"ARG=v8.0.0"* ]]
}

@test "init --version selects and forwards explicit version" {
  make_cached_binary v6.1.0

  run run_wrapper "$WORKSPACE" init project --version v6.1.0
  [ "$status" -eq 0 ]
  [[ "$output" == *"BIN=v6.1.0"* ]]
  [[ "$output" == *"ARG=init"* ]]
  [[ "$output" == *"ARG=project"* ]]
  [[ "$output" == *"ARG=--version"* ]]
  [[ "$output" == *"ARG=v6.1.0"* ]]
}

@test "ccode.yaml without version fails clearly" {
  printf 'ccode_path: ccode\n' >"$WORKSPACE/ccode.yaml"

  run run_wrapper "$WORKSPACE" run task
  [ "$status" -ne 0 ]
  [[ "$output" == *"version is required"* ]]
}
