---
title: Accelerator States
description: State values and metadata exposed by accelerator inspection commands.
---

Accelerator state records are stored per artifact.

```text
<hidden_path>/accelerators/<artifact-id>.accelerated.json
```

## State values

`pending` means the artifact has unresolved generated content or a changed guide.

`resolved` means the output has been changed from the stored generated snapshot or the generated proposal was explicitly accepted with `ccode acc resolve`.

`corrupt` means the state file cannot be parsed or sanitized.

`ambiguous` means state cannot be mapped cleanly to a single artifact.

`missing_artifact` means the target output file is missing.

`missing_guide` means an attached guide file is missing.

## Metadata fields

`ccode acc show <artifactId>` returns metadata as YAML:

```yaml
artifact_id: handlers.acc.go
guide_path: guides/handlers.md
state: pending
```

Use `ccode acc show <artifactId> --json` for JSON metadata. `ccode acc list`
returns a YAML sequence by default; add `--json` for a JSON array. These responses
include the same metadata fields:

```json
{
  "artifact_id": "handlers.acc.go",
  "guide_path": "guides/handlers.md",
  "state": "pending"
}
```

`message` is included when there is a problem to explain. The saved `pending`
flag remains internal; inspection responses expose only the computed `state`.

## Listing behavior

List commands return unresolved resources as a YAML sequence by default:

```bash
ccode acc list
```

Add `--include-resolved` to include resolved resources:

```bash
ccode acc list --include-resolved
```

Add `--json` to either list command when structured JSON output is needed.
An empty list is printed as `[]` in both formats.

Repeated identical state lines are collapsed. Missing artifacts are reported before missing guides. A changed guide file refreshes its saved checksum and marks the artifact pending again.

## Accepting an unchanged proposal

When a pending generated proposal requires no edits, mark it resolved explicitly:

```bash
ccode acc resolve <artifactId>
```

This changes only the accelerator state; it does not modify the output artifact. If a later run changes the generated content or guide, the artifact becomes pending again.

## Inspecting guides and generated content

Use:

```bash
ccode acc guide <artifactId>
ccode acc content <artifactId>
ccode acc diff <artifactId>
```

The guide command prints only the guide's Markdown, without a path or metadata
wrapper. The content command prints the saved generated snapshot. The diff
command compares that snapshot with the current output file. Use these separate
commands when applying and validating an accelerated artifact.
