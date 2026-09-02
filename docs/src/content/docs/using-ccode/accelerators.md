---
title: Accelerators
description: Safe generation for artifacts that need human or agent adjustment.
---

Accelerators are Cohesive Code's adjustment workflow. They let a process propose an artifact, attach instructions, and avoid overwriting a file once it has been changed.

## API

```ts
ctx.accelerate(id, templatePath, data, instructionsPath?);
```

The runtime:

1. Uses the active scope.
2. Renders `templatePath` relative to `ccode_path`.
3. Writes or updates accelerator state.
4. Targets `output_path/<id>`.
5. Writes the target only when it is safe.

## Safe-write behavior

If the target file does not exist, Cohesive Code writes it.

If the target file exists and still matches the last generated snapshot, Cohesive Code may overwrite it with the new rendered content.

If the target file exists and differs from the last generated snapshot, Cohesive Code leaves it alone. This is what protects human and agent adjustments.

## State path

Accelerator state is stored at:

```text
<hidden_path>/accelerators/<scope>/<artifact-id>.accelerated.json
```

The default scope is the process file name without `.ts`. Override it with:

```ts
ctx.setScope("generate-api");
```

## Instructions

An optional instructions file can be attached to an artifact:

```ts
ctx.accelerate(
  "handlers.go",
  "templates/handlers.tpl",
  model,
  "instructions/handlers.md",
);
```

Instruction paths are normalized under `ccode_path`.

## Inspecting work

```bash
ccode list accelerated
ccode list accelerated generate-api
ccode get accelerated generate-api:handlers.go
ccode get accelerated generate-api:handlers.go --instructions
```

Use `--for-agent` for JSON output.

If the generated proposal is already acceptable and needs no edits, mark it adjusted explicitly:

```bash
ccode adjust generate-api:handlers.go
```

This updates only accelerator state. A later change to the generated content or instructions makes the artifact pending again.

## Cleanup behavior

After a successful `ccode run`, the runner removes accelerator state files in scopes used by that run when those states were not produced during the run. This prevents old unresolved state from lingering after a process stops emitting an artifact.
