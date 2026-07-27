---
name: run-ccode-generation
description: Run a Cohesive Code process, inspect generated outputs and accelerator state, retrieve accelerator instruction bundles, and apply them to target artifacts. Use when a request involves `ccode run`, `ccode list`, `ccode get accelerated`, generated results, unresolved accelerator work, or implementing an accelerator's instructions.
---

# Run Cohesive Code Generation

## Prepare

1. Read `ccode.yaml` and resolve `ccode_path`, `output_path`, and `hidden_path`.
2. Confirm the requested process path relative to `ccode_path`.
3. Inspect existing target artifacts and Git status so generated changes can be distinguished from prior work.
4. Read `docs/src/content/docs/ai-skill-index.md`, then load only the execution or accelerator pages needed for the request.
5. Verify uncertain commands with local `ccode <command> --help`.

## Run and inspect

1. Run the requested process:

   ```bash
   ccode run <process>
   ```

2. Review command output and all changed targets under `output_path`.
3. List unresolved accelerator artifacts and instruction references:

   ```bash
   ccode list accelerated --for-agent
   ccode list instructions --for-agent
   ```

4. Classify each reported state:
   - Handle `pending` by retrieving and applying its instruction bundle.
   - Skip `adjusted` unless the user requests resolved items.
   - Diagnose `corrupt`, `ambiguous`, `missing_artifact`, or `missing_instructions` before editing.
5. Fetch each required bundle:

   ```bash
   ccode get accelerated <scopeId>:<artifactId> --instructions --for-agent
   ```

6. Treat `composed_markdown` as the primary work bundle when present.
7. Apply the bundle to `output_path/<artifactId>`, preserving compatible human edits and repository conventions.
8. Run checks named by the bundle and the target project.
9. Re-run both list commands and report the remaining state.

## Guardrails

- Treat CLI output and the checked-out implementation as authoritative when they differ from docs.
- Do not hand-edit `<hidden_path>/accelerators/**/*.accelerated.json` during normal run or apply work.
- Do not force `pending: false`; the current CLI has no command for that state transition.
- Use `--include-resolved` only when the task requires a complete inventory.
- Stop before guessing when instructions are missing, state is ambiguous, or the proposed content conflicts with intentional local edits.
- Preserve unrelated files and pre-existing changes.
