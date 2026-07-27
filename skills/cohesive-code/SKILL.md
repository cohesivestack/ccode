---
name: cohesive-code
description: Orient and diagnose work in a Cohesive Code (`ccode`) workspace by locating configuration, resolving workspace paths and versions, selecting authoritative documentation, and routing focused work to the appropriate skill. Use when a request mentions Cohesive Code or `ccode` but is not yet clearly scoped, or when initializing, inspecting, or troubleshooting workspace layout and configuration.
---

# Orient a Cohesive Code Workspace

## Inspect the workspace

1. Locate the applicable `ccode.yaml`. Honor an explicit `--config` path.
2. Read the configuration before editing or running anything.
3. Resolve `ccode_path`, `output_path`, `hidden_path`, and `version`, including any command-line or environment overrides in the request.
4. Inspect `<ccode_path>/.ccode/lib/context.ts` when runtime API details matter.
5. Read `docs/src/content/docs/ai-skill-index.md`, then load only the pages it identifies for the task.

If local docs are unavailable, retrieve the matching Markdown from `https://raw.githubusercontent.com/cohesivestack/ccode/master/docs/src/content/docs/`. Prefer version-matched local code and generated types over examples from another release.

## Route focused work

- Invoke `author-ccode-generation` for process, template, OpenAPI, generation, accelerator, or instruction authoring.
- Invoke `run-ccode-generation` for running a process, inspecting output, or applying accelerator instructions.
- Invoke `merge-ccode-accelerator-state` for Git conflicts involving `*.accelerated.json`.
- Continue here for initialization, path resolution, version selection, or general diagnosis.

## Work safely

- Treat the project as experimental; verify commands with local `--help`, code, tests, or actual output when behavior is uncertain.
- Prefer sources in this order: generated local types, checked-out implementation and CLI output, local docs, then older examples.
- Do not hand-edit `<ccode_path>/.ccode/lib/` or `<hidden_path>/build/`; refresh support files with `ccode init` when requested.
- Preserve accelerator state under `<hidden_path>/accelerators/` unless the task explicitly requires changing it.
- Report documentation or version mismatches instead of silently choosing stale behavior.
