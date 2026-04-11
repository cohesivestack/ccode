# Accelerator workflow

## Purpose

Use `accelerate` when generation output is a proposal that should be reviewed or adjusted before final adoption.

## Runtime write and tracking behavior

- `ctx.accelerate(id, templatePath, data, instructionsPath?)` renders the template and targets:
  - `output_path/<scope>/<id>`
- The active scope is:
  - process file name by default
  - overridden with `ctx.setScope(...)`
- State is persisted at:
  - `.ccode/state/accelerators.json`

State record fields are stable:

- `id`
- `content` (timestamp + encoded single-line payload)
- `adjusted_at`
- `instructions_path`

## CLI inspection commands

- List pending (not adjusted) artifacts:
  - `ccode list accelerated [scopeId]`
- List instruction references:
  - `ccode list instructions`
- Get one artifact metadata:
  - `ccode get accelerated <scopeId>:<artifactId>`
- Get composed adjustment markdown:
  - `ccode get accelerated <scopeId>:<artifactId> --instructions`
- Read raw instruction markdown:
  - `ccode get instruction <path>`

For machine-readable output:

- add `--for-agent` to any list/get command

## Suggested agent flow

1. Run generator process (`ccode run ...`).
2. Query pending items (`ccode list accelerated --for-agent`).
3. For each item, fetch instruction bundle (`ccode get accelerated ... --instructions --for-agent`).
4. Apply edits to the target file in `output_path/<scope>/<id>`.
5. Marking adjusted is available through the Go API (`MarkAcceleratorAsAdjusted`), and can be integrated by host tooling.

## Guardrails

- Do not decode `content` manually unless building instruction output.
- Do not expose accelerated content in list responses.
- Do not expose instruction markdown from list commands.
- Keep `instructions_path` relative to `ccode_path`.
