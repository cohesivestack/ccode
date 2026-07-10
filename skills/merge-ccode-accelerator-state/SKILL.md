---
name: "merge-ccode-accelerator-state"
description: "Use when resolving Git merge, rebase, cherry-pick, or conflict states involving Cohesive Code accelerator state files under .ccode/accelerators/**/*.accelerated.json, including conflict markers, add/delete conflicts, ambiguous state, missing instructions, and deciding how to preserve accelerator state candidates without rerunning generators during an in-progress Git operation."
---

# Merge Cohesive Code Accelerator State

Use this skill for Git conflict work involving Cohesive Code accelerator state files.

## Documentation Source

Do not duplicate Cohesive Code docs into the skill context. Load only the docs needed for the user request.

1. First look for local docs in `docs/src/content/docs`.
2. If the user names a version, prefer `docs/src/content/docs/<version>` when it exists.
3. If local docs are unavailable, fetch from `https://raw.githubusercontent.com/cohesivestack/ccode/master/docs/src/content/docs/<path>`.
4. If docs and local code disagree, trust the checked-out implementation and tests.

Read these docs as needed:

- Accelerator state format and state values: `reference/accelerator-states.md`
- Accelerator workflow: `using-ccode/accelerators.md`
- CLI inspection commands: `reference/cli.md`
- Workspace paths: `using-ccode/project-layout.md`, `reference/configuration.md`

## Core Rule

During a merge, rebase, cherry-pick, or other in-progress Git operation, preserve accelerator state candidates and remove Git conflict syntax. Do not require `ccode run` to resolve state conflicts, because generator source, templates, specs, or instructions may be temporarily inconsistent.

Prefer ambiguity over destructive guessing. A state file with multiple different valid JSON lines is intentionally inspectable as `ambiguous` later.

## State File Facts

Accelerator state files live at:

```text
<hidden_path>/accelerators/<scope>/<artifact-id>.accelerated.json
```

Each non-empty line should be one JSON state candidate for the same scope/artifact. The CLI can collapse repeated identical lines and report different valid lines as ambiguous.

State fields include:

- `pending`
- `instructions`
- `accelerated_checksum`
- `instructions_checksum`
- `code`

The `code` field is an encoded generated-content snapshot. Do not manually mix `code` from one candidate with checksum fields from another candidate.

## Conflict Workflow

1. Identify conflicted files with Git status.
2. Classify conflicts:
   - Generator inputs: process files, templates, specs, seed data, and instruction markdown.
   - Accelerator state: `.ccode/accelerators/**/*.accelerated.json`.
   - Output artifacts under `output_path`.
3. If generator input files are conflicted, ask whether to resolve those conflicts too. If not, recommend resolving them manually and running this skill again.
4. For accelerator state files, remove Git conflict markers while preserving valid candidate JSON lines.
5. Collapse exact duplicate candidate lines to one line.
6. If multiple different valid candidate lines remain, keep them unless one is clearly invalid or obsolete.
7. Validate that no conflict markers remain in accelerator state files.
8. At the end, recommend rerunning the generator only after the merge/rebase/cherry-pick is complete and the workspace is coherent.

## Candidate Selection Rules

- If conflict markers introduced identical JSON candidates, keep exactly one copy.
- If candidates differ, keep all valid candidates unless a candidate can be safely discarded.
- Safely discard a candidate only when it references an instruction path that is definitely gone and another candidate for the same artifact references an existing instruction path.
- Do not discard a candidate for missing instructions when the instruction file is itself conflicted, renamed, or not yet restored.
- For add/delete conflicts, keep the added state file when unclear.
- Delete a state file only when generator source clearly no longer emits that artifact or the artifact/instruction removal is intentional.

## What Not To Do

- Do not rerun the generator as the first state-conflict resolution step during an in-progress Git operation.
- Do not hand-merge `code`, `accelerated_checksum`, or `instructions_checksum` fields across candidates.
- Do not set `pending: false` merely to quiet the state.
- Do not remove an `instructions` path unless its removal is intentional or clearly superseded.
- Do not leave Git conflict markers in `.accelerated.json` files.

## Final Checks

When Git conflicts are resolved, run inspection commands if practical:

```bash
ccode list accelerated --for-agent
ccode list instructions --for-agent
```

If the workspace is still mid-merge or the generator source was not coherent, finish the Git operation first. Then recommend:

```bash
ccode run <process>
ccode list accelerated --for-agent
ccode list instructions --for-agent
```
