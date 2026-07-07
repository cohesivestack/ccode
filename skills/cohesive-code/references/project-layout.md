# Project layout

## What `ccode init` creates

Running `ccode init [path]` creates a project directory and a config file.

Default result:

```text
./ccode.yaml
./ccode/
  tsconfig.json
  .ccode/
    .gitignore
    build/
    lib/
      context.ts
      openapi.ts
```

If `ccode.yaml` already exists, `init` keeps the file and only adds or updates its top-level `version` entry. `init` refreshes generated support files in `.ccode/lib/`, preserves existing `tsconfig.json`, creates `.ccode/.gitignore` if missing, clears and recreates `.ccode/build/`, and leaves accelerator state untouched.

The default `.ccode/.gitignore` ignores everything inside `.ccode/` except the `.gitignore` file itself, so the repository can keep the hidden folder anchor without committing generated support files, build cache, or runtime state.

## Config and path resolution

Current config keys:

```yaml
ccode_path: ccode
version: v1.2.3
output_path: .
hidden_path: .ccode
```

Notes:

- `path` is still accepted as a legacy alias for `ccode_path`.
- `version` is required in `ccode.yaml` and is used by the wrapper to select the ccode release for the workspace.
- `ccode_path` is resolved relative to the config file directory.
- `output_path` is used as written. Relative values resolve from the directory where the CLI runs.
- `hidden_path` defaults to `.ccode`. When relative, the runtime resolves it under `ccode_path`.
- accelerator state is written to `<hidden_path>/state/accelerators.json`.

Config precedence:

1. Built-in defaults
2. `ccode.yaml`
3. Environment variables
4. CLI flags

Supported environment overrides:

- `CCODE_CCODE_PATH`
- `CCODE_PATH` as a legacy alias
- `CCODE_OUTPUT_PATH`
- `CCODE_HIDDEN_PATH`

Supported CLI overrides:

- `--config`
- `--ccode-path`
- `--output-path`
- `--path` as a deprecated alias for `--ccode-path`

Wrapper version precedence:

1. `--version`
2. nearest `ccode.yaml` `version` entry, or the file passed with `--config`
3. global pin at `~/.config/ccode/version`
4. latest stable GitHub release

## Recommended source layout inside `ccode_path`

The runtime does not force folders beyond the process file path, but this layout keeps projects readable:

```text
ccode/
  api/
    generate.ts
  data/
    seed.json
  specs/
    service.yaml
  templates/
    docs/
      operation.tpl
    sdk/
      client.tpl
```

Recommended conventions:

- Store process entrypoints in a stable folder such as `api/`, `codegen/`, or `processes/`.
- Keep template files under `templates/`.
- Keep OpenAPI files under `specs/`.
- Keep test or seed JSON under `data/`.
- Leave `.ccode/build/` to the compiler cache.
- Treat `.ccode/state/accelerators.json` as runtime state; do not hand-edit it unless explicitly asked.

## Process path rules

The CLI expects `ccode run <process>`, where `<process>` is:

- relative to `ccode_path`
- written without the `.ts` extension
- not absolute
- not `.` or `..`
- not a path that escapes the workspace with `../...`

Example:

- File: `ccode/api/generate.ts`
- Command: `ccode run api/generate`

## Practical agent rules

- Read `ccode.yaml` before creating files.
- Keep templates and spec files inside `ccode_path` unless the user explicitly wants external inputs.
- Do not edit `.ccode/lib/context.ts` in application repos; regenerate it through `ccode init` or change the CLI templates in the source repo instead.
- Run `ccode init` after changing `ccode.yaml` `version` to refresh generated support files for that version.
- For accelerated artifacts, inspect pending items via CLI (`ccode list accelerated`) before editing generated outputs manually.
