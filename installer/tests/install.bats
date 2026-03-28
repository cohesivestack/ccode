#!/usr/bin/env bats

TEST_ROOT=""
TEST_REPO=""
TEST_HOME=""
SOURCE_INSTALL=""
SOURCE_WRAPPER=""

setup() {
  TEST_ROOT="$(mktemp -d "${BATS_TEST_TMPDIR:-/tmp}/install-bats.XXXXXX")"
  TEST_REPO="${TEST_ROOT}/repo"
  TEST_HOME="${TEST_ROOT}/home"
  mkdir -p "${TEST_REPO}/installer/bin" "$TEST_HOME"

  SOURCE_INSTALL="${BATS_TEST_DIRNAME}/../install.sh"
  SOURCE_WRAPPER="${BATS_TEST_DIRNAME}/../bin/ccode"

  cp "$SOURCE_INSTALL" "${TEST_REPO}/installer/install.sh"
  cp "$SOURCE_WRAPPER" "${TEST_REPO}/installer/bin/ccode"
  chmod +x "${TEST_REPO}/installer/install.sh" "${TEST_REPO}/installer/bin/ccode"
}

teardown() {
  rm -rf "$TEST_ROOT"
}

run_installer() {
  local path_value="$1"
  (
    cd "$TEST_REPO"
    HOME="$TEST_HOME" PATH="$path_value" bash installer/install.sh
  )
}

@test "basic install places wrapper at ~/.local/bin/ccode with executable permissions and matching content" {
  [ ! -d "${TEST_HOME}/.local/bin" ]

  run run_installer "/usr/bin:/bin"
  [ "$status" -eq 0 ]

  [ -d "${TEST_HOME}/.local/bin" ]
  [ -f "${TEST_HOME}/.local/bin/ccode" ]
  [ -x "${TEST_HOME}/.local/bin/ccode" ]
  cmp -s "${TEST_REPO}/installer/bin/ccode" "${TEST_HOME}/.local/bin/ccode"
}

@test "reinstall overwrites an existing ~/.local/bin/ccode cleanly" {
  mkdir -p "${TEST_HOME}/.local/bin"
  printf '#!/usr/bin/env bash\necho old-wrapper\n' >"${TEST_HOME}/.local/bin/ccode"
  chmod 0644 "${TEST_HOME}/.local/bin/ccode"

  run run_installer "/usr/bin:/bin"
  [ "$status" -eq 0 ]
  [ -x "${TEST_HOME}/.local/bin/ccode" ]
  cmp -s "${TEST_REPO}/installer/bin/ccode" "${TEST_HOME}/.local/bin/ccode"
}

@test "when ~/.local/bin is in PATH installer reports it is already in PATH without add-hint" {
  mkdir -p "${TEST_HOME}/.local/bin"

  run run_installer "${TEST_HOME}/.local/bin:/usr/bin:/bin"
  [ "$status" -eq 0 ]
  [[ "$output" == *"is already in PATH."* ]]
  [[ "$output" != *"Add it to your shell profile"* ]]
}

@test "when ~/.local/bin is not in PATH installer prints helpful PATH hint" {
  run run_installer "/usr/bin:/bin"
  [ "$status" -eq 0 ]
  [[ "$output" == *"is not currently in PATH."* ]]
  [[ "$output" == *"Add it to your shell profile"* ]]
  [[ "$output" == *"export PATH=\""* ]]
}

@test "installer does not create config version files or release cache directories" {
  run run_installer "/usr/bin:/bin"
  [ "$status" -eq 0 ]

  [ ! -e "${TEST_HOME}/.config/ccode/version" ]
  [ ! -e "${TEST_REPO}/.ccode/version" ]
  [ ! -e "${TEST_HOME}/.cache/ccode/releases" ]
}

@test "installer fails clearly when installer/bin/ccode is missing" {
  rm -f "${TEST_REPO}/installer/bin/ccode"

  run run_installer "/usr/bin:/bin"
  [ "$status" -ne 0 ]
  [[ "$output" == *"wrapper source not found"* ]]
}
