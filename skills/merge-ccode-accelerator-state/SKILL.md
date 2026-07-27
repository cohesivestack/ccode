---
name: merge-ccode-accelerator-state
description: Resolve Git merge, rebase, cherry-pick, and add/delete conflicts in Cohesive Code accelerator state files while preserving valid state candidates. Use when conflicts involve configured `accelerators/**/*.accelerated.json` state, conflict markers, ambiguous accelerator state, or related generator inputs that are temporarily inconsistent during an in-progress Git operation.
---

# Merge Cohesive Code Accelerator State

## Prepare

1. Confirm the active Git operation and inspect `git status`.
2. Read `ccode.yaml` and resolve `ccode_path`, `output_path`, and `hidden_path`.
3. Read `docs/src/content/docs/reference/accelerator-states.md` and `docs/src/content/docs/using-ccode/accelerators.md` when available.
4. Separate conflicts into generator inputs, output artifacts, and `<hidden_path>/accelerators/**/*.accelerated.json`.
5. Keep user changes and unrelated conflicts out of scope unless the user asks to resolve them.

Do not run a generator while its process, templates, specs, inputs, or instructions are conflicted. Preserve recoverable state until the Git operation and generator inputs are coherent.

## Resolve state files

1. Treat every non-empty line as a candidate for the scope and artifact encoded by the file path.
2. Remove Git conflict markers and blank lines.
3. Keep one copy of exact duplicate candidate lines.
4. Keep multiple distinct valid JSON candidates when no candidate is clearly obsolete. The CLI will report the file as `ambiguous` for later repair.
5. Preserve each candidate as a complete record. Never combine `code`, `accelerated_checksum`, `instructions_checksum`, `instructions`, or `pending` values from different candidates.
6. Prefer keeping an added state file in an unclear add/delete conflict.
7. Delete a state file only when the merged generator intentionally stops emitting that artifact or the user confirms the removal.
8. Discard a candidate only when its removal is demonstrably intentional or another candidate clearly supersedes it. Do not infer obsolescence from a currently conflicted or renamed instruction file.

## Verify

1. Confirm that state files contain no conflict markers.
2. Validate each retained non-empty line as a complete JSON object.
3. Confirm that every edited state path still identifies the intended scope and artifact.
4. Inspect state only after the related conflicts are resolved:

   ```bash
   ccode list accelerated --for-agent
   ccode list instructions --for-agent
   ```

5. After the Git operation completes and generator inputs are coherent, run the affected process to rebuild or disambiguate state:

   ```bash
   ccode run <process>
   ccode list accelerated --for-agent
   ccode list instructions --for-agent
   ```

6. Report retained ambiguity and any missing artifact or instruction instead of hiding it with `pending: false`.
