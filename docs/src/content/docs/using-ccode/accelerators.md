---
title: Accelerators
description: Safe generation for artifacts that need human or agent resolvement.
---

Accelerators are Cohesive Code's resolvement workflow. They let a process propose an artifact, attach a guide, and avoid overwriting a file once it has been changed.

## API

```ts
ctx.accelerate(id, templatePath, data, guidePath?);
```

The runtime:

1. Renders `templatePath` relative to `ccode_path`.
2. Writes or updates accelerator state.
3. Targets `output_path/<id>`.
4. Writes the target only when it is safe.

## Safe-write behavior

If the target file does not exist, Cohesive Code writes it.

If the target file exists and still matches the last generated snapshot, Cohesive Code may overwrite it with the new rendered content.

If the target file exists and differs from the last generated snapshot, Cohesive Code leaves it alone. This is what protects human and agent resolvements.

## State path

Accelerator state is stored at:

```text
<hidden_path>/accelerators/<artifact-id>.accelerated.json
```

Artifact IDs are output paths relative to `output_path` and must be unique across
the workspace. Nested IDs create matching directories below both `output_path`
and the accelerator state directory.

## Guides

An optional guide file can be attached to an artifact:

```ts
ctx.accelerate(
  "handlers.acc.go",
  "templates/handlers.tpl",
  model,
  "guides/handlers.md",
);
```

Guide paths are normalized under `ccode_path`.

## Inspecting work

```bash
ccode acc list
ccode acc show handlers.acc.go
ccode acc show handlers.acc.go --guide
```

Use `--for-agent` for JSON output.

To print only the original generated content saved in accelerator state:

```bash
ccode acc content handlers.acc.go
```

This returns the saved snapshot from the latest generation, including when the output file has been edited.

If the generated proposal is already acceptable and needs no edits, mark it resolved explicitly:

```bash
ccode acc resolve handlers.acc.go
```

This updates only accelerator state. A later change to the generated content or guide makes the artifact pending again.

## Removing obsolete artifacts

Use [Pruning](/using-ccode/prune/) after acceleration to remove obsolete `.acc.*`
outputs and their state. Artifacts emitted during the current run stay protected,
including edited files that `accelerate` leaves in place.
