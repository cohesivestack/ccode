---
name: ccode-produce
description: Run an existing Cohesive Code (ccode) process in an initialized project, review generated outputs, then hand off pending accelerated artifacts to ccode-resolve. Use ccode-resolve directly when generation has already completed.
---

# Produce Cohesive Code

Use the bundled Python 3 helper for preflight checks, process selection, and
execution, then hand off guide-driven edits and artifact resolution to `ccode-resolve`.
The helper uses only the standard library and invokes `ccode` without a shell.
On Windows use `python` or `py -3` if `python3` is unavailable; `ccode` must be
installed and runnable on that platform.

## Establish the project

Work from the intended project's working directory. Read its `ccode.yaml` (or the
user's explicit config), applicable repository instructions, and existing changes.
If no config exists, stop and tell the user the project needs initialization.
Do not initialize it or author processes, templates, or guides in this workflow.

Resolve effective paths before invoking the helper:

- `ccode_path`: defaults to `ccode`, relative to the config's directory.
- `output_path`: defaults to `.`, relative to the working directory.
- `hidden_path`: defaults to `.ccode`, relative to the effective `ccode_path`.
- Honor `CCODE_CCODE_PATH`, `CCODE_OUTPUT_PATH`, and `CCODE_HIDDEN_PATH` over config
  values, then explicit user path overrides over those environment values. Relative
  `ccode_path` overrides are relative to the working directory, unlike config values.

Pass absolute effective paths. The agent reads YAML so the dependency-free helper
does not need to implement a subset of YAML. The helper also requires the source
directory, `tsconfig.json`, and `<hidden_path>/lib/context.ts` as initialization
evidence. If these are missing, stop and report what is missing; do not run `init`.
The CLI validates the config (including its required version) when invoked.

## Run generation

Replace `SKILL_DIR` with this skill's actual directory:

```bash
python3 "SKILL_DIR/scripts/produce.py" \
  --config "/project/ccode.yaml" \
  --ccode-path "/project/ccode" \
  --output-path "/project" \
  --hidden-path "/project/ccode/.ccode"
```

Add `--process "name"` when the user supplied a process name, including a prior
selection. Otherwise the helper calls `ccode proc list` and runs exactly `main`
when present. If it returns `selection_required`, show **all** available processes
and ask the user to select one, even if there is only one. Resume with `--process`
after their answer. If no processes exist, report that and stop.

The helper stops immediately on a process execution error. Report the error;
do not invoke `ccode-resolve`, retry the process, or change generation inputs as
part of handling that failure.

stdout is a JSON report; command logs go to stderr. Exit code 1 means failure.
Exit code 0 requires checking `status`: only `generated` means the process ran
successfully. The report includes `process`, `next_skill: "ccode-resolve"`, and
the absolute `config`, `ccode_path`, `hidden_path`, and `output_path` needed by the
resolver. Its `command_prefix` and `environment` preserve follow-up CLI settings.

## Hand off resolution

After `generated`, inspect command output and the changed generated targets, then
use the installed `ccode-resolve` skill with the **same working directory, config,
and effective paths** from the report. That skill collects existing pending
artifacts, applies their guides, validates edits, and resolves completed work.
Do not rerun generation to refresh accelerator status; the resolver owns that
inspection. Resolution instructions live in that skill rather than this one.

If `ccode-resolve` is unavailable, report that generation completed and resolution
is still needed. Identify the missing companion skill and its installation command:

```bash
npx skills add cohesivestack/ccode --skill ccode-resolve
```

Do not treat a successful generation report as evidence that artifacts were
resolved. Finish by reporting the process, generated changes, checks, and the
resolver's completed and remaining work, or the missing-skill handoff.

Preserve pre-existing work and generated support files. Initialization and changes
to generation inputs belong to `ccode-author`.
