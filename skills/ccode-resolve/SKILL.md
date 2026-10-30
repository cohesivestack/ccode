---
name: ccode-resolve
description: Apply guides to existing pending Cohesive Code (ccode) accelerated artifacts, validate output changes, and resolve completed artifacts without running generation. Use after a production run or whenever pending artifacts already exist; use ccode-produce when generation is needed first.
---

# Resolve Cohesive Code

Use the bundled Python 3 helper to collect pending artifact metadata, guides,
diffs, and saved generated snapshots. Use the agent to interpret guides, edit
output files, validate them, and resolve completed artifacts. The helper uses only
the standard library and invokes `ccode` without a shell. On Windows use `python`
or `py -3` if `python3` is unavailable; `ccode` must be installed and runnable.

## Establish the project

Work from the intended project's working directory. Read `ccode.yaml` (or the
user's explicit config), applicable repository instructions, and existing changes.
If invoked after `ccode-produce`, retain its working directory, config, and
absolute effective paths. If no config exists, report the missing config.

Resolve paths before invoking the helper:

- `ccode_path`: defaults to `ccode`, relative to the config's directory.
- `output_path`: defaults to `.`, relative to the working directory.
- `hidden_path`: defaults to `.ccode`, relative to the effective `ccode_path`.
- Honor `CCODE_CCODE_PATH`, `CCODE_OUTPUT_PATH`, and `CCODE_HIDDEN_PATH` over config
  values, then explicit user path overrides over those environment values. Relative
  `ccode_path` overrides resolve from the working directory.

Pass absolute effective paths. The agent reads YAML; the dependency-free helper
leaves config validation to the CLI. Resolution uses existing state and outputs;
it does not require `tsconfig.json` or generated TypeScript support modules.
Do not initialize a project, discover or run processes, or change generation
inputs in this workflow.

## Collect pending work

Replace `SKILL_DIR` with this skill's actual directory:

```bash
python3 "SKILL_DIR/scripts/resolve.py" \
  --config "/project/ccode.yaml" \
  --ccode-path "/project/ccode" \
  --output-path "/project" \
  --hidden-path "/project/ccode/.ccode"
```

The helper calls `ccode acc list --json`, skips resolved artifacts, and collects
each usable pending artifact's guide, diff, and snapshot with:

```bash
ccode acc guide <artifactId>
ccode acc diff <artifactId>
ccode acc content <artifactId>
```

The guide command returns only Markdown, without a path or JSON wrapper. For
individual metadata use `ccode acc show <artifactId> --json`; list and show
otherwise default to YAML. stdout is a JSON report; command logs go to stderr.
Exit code 1 means failure. Exit code 0 still requires checking the report's
`status` and artifact statuses. `no_pending_artifacts` means there is no work.

The report's `command_prefix` and `environment` carry the configuration for
follow-up CLI calls. Set that environment for each call without changing the
parent shell permanently. A collection failure is not an empty work list.

## Apply, validate, and resolve

1. For every artifact marked `ready`, read its current file under `output_path`
   and its complete guide, diff, and generated snapshot. If the report was
   truncated, retrieve the individual fields with the CLI before editing.
2. Apply the guide's required changes while preserving compatible human edits
   and repository conventions. The diff runs **from generated snapshot to current
   file**: `-` is generated content and `+` is current content. Do not apply it
   blindly as a patch. An empty diff does not mean the guide requires no work.
3. Leave `needs_guide` artifacts pending, record their IDs and guide paths when
   available, and continue independent work. Do not invent guides. For `blocked`
   artifacts (corrupt or ambiguous state, missing output, or inspection failure),
   leave them unresolved and record the reason.
4. Run the guide's applicable validation and relevant project checks. Resolve only
   artifacts whose required edits and checks succeeded, including proposals that
   need no changes. If a shared check fails and its impact cannot be isolated,
   keep affected artifacts pending and report the failure.

   ```bash
   ccode acc resolve <artifactId>
   ```

   Use the report's command prefix and environment. Do not edit accelerator state
   by hand or resolve skipped, failed, or unvalidated artifacts.
5. Run the same helper again with the same path options to refresh remaining work.
   Every invocation inspects existing artifacts without running generation.
   Report changed files, checks, resolved artifacts, and remaining blockers.
   Explicitly list every artifact that still requires a guide.

Preserve pre-existing changes and generated support files. Report state repairs
or input changes that are needed rather than performing them automatically.
