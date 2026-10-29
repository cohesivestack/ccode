---
name: ccode-author
description: Initialize Cohesive Code (ccode) projects and author or modify their configuration, TypeScript processes, Gonja templates, input models, and accelerator guides. Use for creating or maintaining generation inputs; use ccode-produce for running existing processes and applying accelerated artifacts.
---

# Author Cohesive Code

Use the bundled Python 3 helper for inspection, initialization, process scaffolding,
and creating template and guide files. Use the agent for the actual generation
logic and prose. The helper uses only the standard library, invokes `ccode` without
a shell, and refuses to overwrite authored files. On Windows use `python` or
`py -3` if `python3` is unavailable; `ccode` must be installed and runnable.

## Establish the project

Work from the intended project's working directory. Read its config, applicable
repository instructions, existing source, target conventions, and Git changes.
Honor an explicit config path; otherwise use `ccode.yaml` in that directory.

Resolve the effective paths and pass them as absolute paths to the helper:

- `ccode_path`: defaults to `ccode`, relative to the config's directory.
- `output_path`: defaults to `.`, relative to the working directory.
- `hidden_path`: defaults to `.ccode`, relative to the effective `ccode_path`.
- Runtime environment overrides (`CCODE_CCODE_PATH`, `CCODE_OUTPUT_PATH`, and
  `CCODE_HIDDEN_PATH`) take precedence over config, followed by explicit user path
  overrides. Relative source overrides resolve from the working directory.

The agent reads YAML; the dependency-free helper does not implement a partial YAML
parser. Replace `SKILL_DIR` with this skill's actual directory:

```bash
python3 "SKILL_DIR/scripts/author.py" \
  --config "/project/ccode.yaml" \
  --ccode-path "/project/ccode" \
  --output-path "/project" \
  --hidden-path "/project/ccode/.ccode" inspect
```

Reuse these path options before each subcommand below. stdout is a JSON report;
command logs go to stderr. Exit code 1 means failure. Exit code 0 still requires
checking `status`, including `initialization_required`. Inspection calls
`ccode proc list` only after finding config, source directory, `tsconfig.json`, and
`<hidden_path>/lib/context.ts`.

## Initialize when needed

For a new project or incomplete support files, use the helper's `init --version`
subcommand. Keep an existing config's version unless the user requested a change.
For a new project, use the user's chosen version; if none is established, ask
which version to use before initialization. Do not adopt an example version or a
moving `latest` default.

```bash
python3 "SKILL_DIR/scripts/author.py" \
  --config "/project/ccode.yaml" \
  --ccode-path "/project/ccode" \
  --output-path "/project" \
  --hidden-path "/project/ccode/.ccode" init --version "vX.Y.Z"
```

For a new config the helper writes the selected paths as quoted YAML scalars,
then invokes `ccode init`. It preserves existing config text; the CLI updates its
version and refreshes support files, preserves authored files and accelerator
state, and clears the build cache. A failed initialization may leave config or
support files in place: report the error and inspect before attempting a fix.

**Existing config:** native `ccode init` takes source and hidden paths from the
config, ignoring runtime path overrides. For this subcommand pass the paths
resolved from that config. Change config paths only when requested, before init,
and then inspect using the intended runtime overrides. Never claim an override
directory was initialized simply because the CLI succeeded.

If `hidden_path` is customized, inspect the `@ccode/*` mapping in `tsconfig.json`;
the CLI's starter points to `./.ccode/lib/*`. Update the workspace alias to the
actual support directory when necessary, preserving unrelated compiler settings.
Do not hand-edit generated API files under `<hidden_path>/lib/` or build files.

## Author processes, templates, and guides

Read [references/authoring.md](references/authoring.md) before writing generation
logic. The inspection report lists source files, processes, and support modules.
Read the installed `<hidden_path>/lib/context.ts` and relevant public declarations
for the project's version instead of guessing signatures or importing `internal/`.

Use the helper to create only the files needed by the request:

```text
add-process api/generate
add-template templates/handlers.tpl --content-file /temporary/handlers.tpl
add-guide guides/handlers.md --content-file /temporary/handlers.md
```

These are subcommands appended after the shared path options. `add-process` calls
`ccode proc add` for a default-exported `Context` function; edit that scaffold to
implement the requested logic. For templates and guides, first author the complete
UTF-8 content in a temporary file, then pass it with `--content-file`. The helper
copies its bytes unchanged, rejects empty content, path escapes and generated
support destinations, and refuses existing destinations. For existing files,
inspect and edit them directly rather than trying to recreate or overwrite them.

Keep generation inputs under `ccode_path`. Use stable relative output paths and
workspace-unique accelerator IDs. Write a guide whenever downstream adaptation is
required, describing the changes, preserved behavior, and validation. Do not add
empty guides or placeholder templates merely to satisfy a scaffold.

## Validate and hand off

Inspect the completed source and relevant `tsconfig.json` settings, then rerun
`inspect` to verify process discovery. Run applicable existing authoring checks.
The helper's inventory does not type-check TypeScript or render templates.

When execution is needed to verify the requested authoring work, use
`ccode-produce` with the **exact changed process**. That skill owns generation,
guide-driven output changes, validation, and artifact resolution. If it is not
installed, use `ccode run <process>` with the same path options/environment,
inspect the resulting outputs and `ccode acc list`, and report pending artifacts
for production follow-up. Do not resolve them as part of authoring alone.

Report initialized or changed inputs, checks performed, and whether the process
was executed. Preserve unrelated work and accelerator state. Do not include live
database credentials in source, templates, guides, or reports.
