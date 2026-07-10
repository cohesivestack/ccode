---
name: "cohesive-code"
description: "Use when working with Cohesive Code or ccode projects, including workspace detection, initialization, authoring TypeScript generation processes, running generators, inspecting accelerated artifacts, and understanding accelerator state. Prefer the specialized skills author-ccode-generation, run-ccode-generation, and merge-ccode-accelerator-state for focused authoring, execution, or Git conflict tasks."
---

# Cohesive Code

Use this skill as the broad Cohesive Code workspace guide. For focused tasks, prefer:

- `author-ccode-generation` for creating or changing processes, templates, OpenAPI workflows, generated artifacts, accelerated artifacts, and instruction markdown.
- `run-ccode-generation` for running `ccode run`, inspecting accelerator output, and applying accelerator instruction bundles.
- `merge-ccode-accelerator-state` for merge, rebase, cherry-pick, or conflict work involving `.ccode/accelerators/**/*.accelerated.json`.

## Documentation Source

Do not duplicate Cohesive Code docs into the skill context. Load only the docs needed for the user request.

1. First look for local docs in the current workspace:
   - `docs/src/content/docs`
   - For older versions: `docs/src/content/docs/<version>`
2. If local docs are unavailable, fetch the matching Markdown file from:
   - `https://raw.githubusercontent.com/cohesivestack/ccode/master/docs/src/content/docs/<path>`
3. If the user names a version, prefer the versioned docs folder when it exists.
4. If docs and local code disagree, trust the checked-out code, generated context types, tests, and actual CLI output, then mention the mismatch.

Use this routing table to choose the smallest relevant docs set:

- Workspace overview and agent map: `ai-skill-index.md`
- Setup and first run: `getting-started.md`
- Project paths and initialized layout: `using-ccode/project-layout.md`
- Configuration: `reference/configuration.md`
- Wrapper behavior and pinned versions: `using-ccode/wrapper-and-versions.md`
- Process contract: `using-ccode/processes.md`
- Templates: `using-ccode/templates.md`
- Runtime API: `reference/runtime-api.md`
- Standard generation: `using-ccode/generation.md`
- OpenAPI generation: `using-ccode/openapi-workflows.md`, `cookbook/openapi-docs.md`
- Accelerators: `using-ccode/accelerators.md`, `reference/accelerator-states.md`, `cookbook/accelerated-artifact.md`
- CLI commands: `reference/cli.md`
- Minimal examples: `cookbook/minimal-process.md`, `cookbook/accelerated-artifact.md`, `cookbook/openapi-docs.md`

## Working Pattern

When working in a Cohesive Code workspace:

1. Detect `ccode.yaml` and read it before editing files.
2. Resolve `ccode_path`, `output_path`, and `hidden_path` from config or CLI flags.
3. Read the smallest relevant docs from the routing table.
4. Prefer the generated local context contract at `<ccode_path>/.ccode/lib/context.ts` over stale examples.
5. Do not hand-edit `<ccode_path>/.ccode/lib/*` or `.ccode/build/*` in application workspaces.
6. Use exact CLI output to drive fixes when validation is practical.

## Core Conventions

- Keep process source, templates, specs, seed data, and accelerator instructions under `ccode_path`.
- Invoke a process as `ccode run <relative/process/path-without-.ts>`.
- Use `ctx.generate(...)` for artifacts that should be overwritten by generation.
- Use `ctx.accelerate(...)` for artifacts that are generated proposals and may need human or agent adjustment.
- Inspect unresolved accelerator work with `ccode list accelerated --for-agent`.
- Fetch an accelerator instruction bundle with `ccode get accelerated <scopeId>:<artifactId> --instructions --for-agent`.
