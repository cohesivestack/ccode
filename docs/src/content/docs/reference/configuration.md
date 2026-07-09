---
title: Configuration
description: ccode.yaml keys, defaults, and override precedence.
---

`ccode.yaml` is the workspace contract.

## Example

```yaml
ccode_path: ccode
version: v0.1.0
output_path: .
hidden_path: .ccode
```

## Keys

`ccode_path` is where process source, templates, specs, data, instructions, and generated support files live. Default: `ccode`.

`version` is required. The wrapper uses it to choose the release binary for the workspace.

`output_path` is the root path where generated artifacts are written. Default: `.`.

`hidden_path` is the internal state/build folder. Default: `.ccode`.

## Resolution

- `ccode_path` is resolved relative to the config file directory.
- `output_path` is used as written. Relative values resolve from the directory where the CLI runs.
- `hidden_path` defaults to `.ccode`. When relative, the runtime resolves it under `ccode_path`.

## Precedence

Runtime config values are assembled in this order:

1. Built-in defaults
2. `ccode.yaml`
3. Environment variables
4. CLI flags

## Environment overrides

```bash
CCODE_CCODE_PATH=ccode
CCODE_OUTPUT_PATH=.
CCODE_HIDDEN_PATH=.ccode
```

## CLI overrides

```bash
ccode --config ./ccode.yaml run api/generate
ccode --ccode-path ./generators run api/generate
ccode --output-path ./out run api/generate
```
