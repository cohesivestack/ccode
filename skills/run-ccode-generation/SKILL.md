---
name: "run-ccode-generation"
description: "Use when running Cohesive Code generator processes and inspecting results, including ccode run, generated output files, accelerated artifacts, accelerator instructions, ccode list accelerated, ccode list instructions, ccode get accelerated --instructions, and applying instruction bundles to target artifacts."
---

# Run Cohesive Code Generation

Use this skill for the operator stage: run a generator, inspect what it produced, fetch accelerator instruction bundles, and apply the requested adjustments to generated proposal artifacts.

## Experimental Project Handling

Cohesive Code is experimental and changes frequently. Continue the requested run or inspection workflow, but verify behavior against local docs, actual CLI output, and generated state. Do not assume commands, output shapes, or accelerator states are stable.

## Documentation Source

Do not duplicate Cohesive Code docs into the skill context. Load only the docs needed for the user request.

1. First look for local docs in `docs/src/content/docs`.
2. If the user names a version, prefer `docs/src/content/docs/<version>` when it exists.
3. If local docs are unavailable, fetch from `https://raw.githubusercontent.com/cohesivestack/ccode/master/docs/src/content/docs/<path>`.
4. If docs and CLI behavior disagree, trust actual command output and the checked-out implementation.

Read these docs as needed:

- CLI commands: `reference/cli.md`
- Accelerator workflow: `using-ccode/accelerators.md`
- Accelerator states: `reference/accelerator-states.md`
- Runtime API background: `reference/runtime-api.md`
- Project paths: `using-ccode/project-layout.md`, `reference/configuration.md`
- Applied examples: `cookbook/accelerated-artifact.md`, `cookbook/minimal-process.md`

## Run Pattern

1. Read `ccode.yaml` and identify `ccode_path`, `output_path`, and `hidden_path`.
2. Run the requested process:

   ```bash
   ccode run <process>
   ```

3. Inspect unresolved accelerator artifacts:

   ```bash
   ccode list accelerated --for-agent
   ```

4. Inspect instruction references:

   ```bash
   ccode list instructions --for-agent
   ```

5. For each unresolved accelerator item, fetch the instruction bundle:

   ```bash
   ccode get accelerated <scopeId>:<artifactId> --instructions --for-agent
   ```

6. Apply the instruction bundle to the target artifact under `output_path`.
7. Re-list accelerator state to confirm remaining work and problems.

## Accelerator State Handling

- Treat `pending` as unresolved generated content or changed instructions.
- Treat `adjusted` as resolved unless the user asks to include resolved items.
- Treat `corrupt`, `ambiguous`, `missing_artifact`, and `missing_instructions` as problems that need attention.
- Use `--include-resolved` only when the task needs a full inventory.
- Do not edit `.ccode/accelerators/**/*.accelerated.json` during normal run/apply work unless the task is specifically about state repair or conflict resolution.

## Applying Instructions

- Use `composed_markdown` from `--for-agent` output as the main work bundle when available.
- Preserve existing human edits in the target artifact unless instructions explicitly replace them.
- If generated content and existing artifact differ, merge intentionally instead of blindly overwriting.
- If instructions are missing or state is ambiguous, report that and avoid guessing beyond clearly safe edits.
