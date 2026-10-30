---
name: ccode-produce
description: Run an existing Cohesive Code (ccode) process, apply pending accelerated artifacts using their guides and diffs, validate changes, and resolve completed artifacts. Use for production runs in initialized projects; initialization and template or process authoring are outside this skill.
---

# Produce Cohesive Code

Use the bundled Python 3 helper for preflight checks, process selection, execution,
and collecting accelerator work. Use the agent to interpret guides and edit files.
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

## Produce and collect

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

The helper stops immediately on a process execution error; do not apply or resolve
accelerators after a failed run. After success it calls `ccode acc list`, then
collects each usable pending artifact's guide, diff, and generated snapshot via:

```bash
ccode acc guide <artifactId>
ccode acc diff <artifactId>
ccode acc content <artifactId>
```

The current guide command is `acc guide`, not `get guide`. The report's
`command_prefix` and `environment` carry the same configuration for follow-up CLI
calls. Set the reported environment for each call; do not change the parent shell
permanently. stdout is a JSON report; command logs go to stderr. Exit code 1 means
failure, while exit code 0 requires checking the report's `status` and artifacts.

## Apply, validate, and resolve

1. Process every artifact marked `ready` in the report. Read its current file under
   `output_path` and its complete guide, diff, and generated snapshot. If tool output
   was truncated, retrieve the individual fields with the CLI before editing.
2. Apply the changes required by the guide, preserving compatible human edits and
   repository conventions. The diff runs **from generated snapshot to current
   file**: `-` is generated content and `+` is current content. Do not blindly apply
   it as a patch. An empty diff does not mean the guide requires no work.
3. For `needs_guide`, leave the artifact pending, record its ID and guide path when
   available, and continue with other artifacts. Do not invent a guide. For
   `blocked` (including corrupt or ambiguous state, missing output, or inspection
   failure), leave it unresolved, record the reason, and continue independent work.
4. Run the guide's applicable validation and relevant project checks. Resolve only
   artifacts whose required edits and checks succeeded, including proposals that
   need no changes. If a shared check fails and its impact cannot be isolated, keep
   the affected artifacts pending and report the failure.

   ```bash
   ccode acc resolve <artifactId>
   ```

   Use the report's command prefix and environment. Do not edit accelerator state
   by hand or resolve skipped, failed, or unvalidated artifacts.
5. Invoke the helper with the same path options plus `--inspect-only` to refresh
   remaining work **without rerunning the process**. Report the process executed,
   changes, checks, resolved artifacts, and remaining blockers. Explicitly list
   every artifact that requires a guide, even if others were completed.

Preserve pre-existing changes and generated support files. Do not rerun a failing
process, change generation inputs, or repair accelerator state automatically.

