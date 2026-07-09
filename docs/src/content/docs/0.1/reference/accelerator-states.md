---
title: Accelerator States
description: State values and metadata exposed by accelerator inspection commands.
slug: 0.1/reference/accelerator-states
---

Accelerator state records are stored per scope and artifact.

```text
<hidden_path>/accelerators/<scope>/<artifact-id>.accelerated.json
```

## State values

`pending` means the artifact has unresolved generated content or changed instructions.

`adjusted` means the output has been changed from the stored generated snapshot.

`corrupt` means the state file cannot be parsed or sanitized.

`ambiguous` means state cannot be mapped cleanly to a single artifact.

`missing_artifact` means the target output file is missing.

`missing_instructions` means an attached instruction file is missing.

## Metadata fields

Inspection responses include:

```json
{
  "scope_id": "generate-api",
  "artifact_id": "handlers.go",
  "instructions_path": "instructions/handlers.md",
  "pending": true,
  "state": "pending",
  "message": ""
}
```

## Listing behavior

List commands return unresolved resources by default:

```bash
ccode list accelerated
ccode list instructions
```

Add `--include-resolved` to include adjusted resources:

```bash
ccode list accelerated --include-resolved
```

Repeated identical state lines are collapsed. Missing artifacts are reported before missing instructions. Changed instruction files refresh their saved checksum and mark the artifact pending again.

## Adjustment bundle

Use:

```bash
ccode get accelerated <scopeId>:<artifactId> --instructions --for-agent
```

The JSON response includes instruction markdown, decoded accelerated content, and a composed markdown bundle suitable for an agent prompt.
