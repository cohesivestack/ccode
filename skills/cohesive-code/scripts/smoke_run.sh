#!/usr/bin/env sh
set -eu

if [ "$#" -lt 1 ]; then
  echo "usage: sh skills/cohesive-code/scripts/smoke_run.sh <process>" >&2
  echo "optional env: CCODE_RUNNER='ccode' or CCODE_RUNNER='go run .'" >&2
  echo "optional env: CCODE_CONFIG=path/to/ccode.yaml" >&2
  exit 1
fi

PROCESS="$1"
RUNNER="${CCODE_RUNNER:-ccode}"
CONFIG="${CCODE_CONFIG:-}"
EXTRA_ARGS="${CCODE_EXTRA_ARGS:-}"

if [ -n "$CONFIG" ]; then
  CMD="$RUNNER --config \"$CONFIG\" run \"$PROCESS\""
else
  CMD="$RUNNER run \"$PROCESS\""
fi

if [ -n "$EXTRA_ARGS" ]; then
  CMD="$CMD $EXTRA_ARGS"
fi

echo "+ $CMD"
exec sh -lc "$CMD"
