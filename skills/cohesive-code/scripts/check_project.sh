#!/usr/bin/env sh
set -eu

ROOT="${1:-.}"
CONFIG="${CCODE_CONFIG:-$ROOT/ccode.yaml}"

if [ ! -f "$CONFIG" ]; then
  echo "config file not found: $CONFIG" >&2
  exit 1
fi

trim_quotes() {
  printf "%s" "$1" | sed "s/^['\"]//; s/['\"]$//"
}

read_yaml_value() {
  key="$1"
  file="$2"
  value="$(awk -F: -v search_key="$key" '
    $1 ~ "^[[:space:]]*" search_key "[[:space:]]*$" {
      sub(/^[^:]*:[[:space:]]*/, "", $0)
      print $0
      exit
    }
  ' "$file")"
  trim_quotes "$value"
}

CONFIG_DIR="$(cd "$(dirname "$CONFIG")" && pwd -P)"
CCODE_PATH="$(read_yaml_value "ccode_path" "$CONFIG")"
if [ -z "$CCODE_PATH" ]; then
  CCODE_PATH="$(read_yaml_value "path" "$CONFIG")"
fi
OUTPUT_PATH="$(read_yaml_value "output_path" "$CONFIG")"
HIDDEN_PATH="$(read_yaml_value "hidden_path" "$CONFIG")"

[ -n "$CCODE_PATH" ] || CCODE_PATH="ccode"
[ -n "$OUTPUT_PATH" ] || OUTPUT_PATH="."
[ -n "$HIDDEN_PATH" ] || HIDDEN_PATH=".ccode"

case "$CCODE_PATH" in
  /*) CCODE_ABS="$CCODE_PATH" ;;
  *) CCODE_ABS="$CONFIG_DIR/$CCODE_PATH" ;;
esac

case "$HIDDEN_PATH" in
  /*) HIDDEN_ABS="$HIDDEN_PATH" ;;
  *) HIDDEN_ABS="$CCODE_ABS/$HIDDEN_PATH" ;;
esac

echo "config: $CONFIG"
echo "config_dir: $CONFIG_DIR"
echo "ccode_path: $CCODE_PATH"
echo "output_path: $OUTPUT_PATH"
echo "hidden_path: $HIDDEN_PATH"
echo "resolved_ccode_path: $CCODE_ABS"
echo "resolved_hidden_path: $HIDDEN_ABS"

if [ -d "$CCODE_ABS" ]; then
  echo "workspace_dir: present"
else
  echo "workspace_dir: missing"
fi

if [ -f "$CCODE_ABS/tsconfig.json" ]; then
  echo "tsconfig: present"
else
  echo "tsconfig: missing"
fi

if [ -f "$HIDDEN_ABS/lib/context.ts" ]; then
  echo "context_types: present"
else
  echo "context_types: missing"
fi

if [ -f "$HIDDEN_ABS/lib/openapi.ts" ]; then
  echo "openapi_types: present"
else
  echo "openapi_types: missing"
fi

if [ -d "$HIDDEN_ABS/build" ]; then
  echo "build_cache: present"
else
  echo "build_cache: missing"
fi

if [ -f "$HIDDEN_ABS/state/accelerators.json" ]; then
  echo "accelerator_state: present"
else
  echo "accelerator_state: missing"
fi

if [ -d "$CCODE_ABS" ]; then
  TS_COUNT="$(find "$CCODE_ABS" -type f -name '*.ts' ! -path "$HIDDEN_ABS/*" | wc -l | tr -d ' ')"
  echo "typescript_files: $TS_COUNT"
fi
