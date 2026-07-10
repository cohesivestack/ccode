---
name: "author-ccode-generation"
description: "Use when authoring or changing Cohesive Code generator source: TypeScript process files, OpenAPI schemas/spec inputs, Gonja templates, ctx.generate outputs, ctx.accelerate accelerated artifacts, scopes, and accelerator instruction markdown. Use for programmers designing generator and accelerator workflows before running them."
---

# Author Cohesive Code Generation

Use this skill when creating or changing the generator side of a Cohesive Code workspace: process code, templates, specs, generated artifacts, accelerated artifacts, scopes, and instruction markdown.

## Experimental Project Handling

Cohesive Code is experimental and changes frequently. Continue the requested authoring workflow, but verify behavior against local docs, generated context types, actual CLI output, and tests. Do not assume APIs, templates, or accelerator semantics are stable.

## Documentation Source

Do not duplicate Cohesive Code docs into the skill context. Load only the docs needed for the user request.

1. First look for local docs in `docs/src/content/docs`.
2. If the user names a version, prefer `docs/src/content/docs/<version>` when it exists.
3. If local docs are unavailable, fetch from `https://raw.githubusercontent.com/cohesivestack/ccode/master/docs/src/content/docs/<path>`.
4. If docs and local source disagree, trust the checked-out code, generated context types, tests, and actual CLI output.

Read these docs as needed:

- Workspace and source-of-truth map: `ai-skill-index.md`
- Process contract: `using-ccode/processes.md`
- Templates: `using-ccode/templates.md`
- Runtime API: `reference/runtime-api.md`
- Standard generation: `using-ccode/generation.md`
- OpenAPI generation: `using-ccode/openapi-workflows.md`, `cookbook/openapi-docs.md`
- Accelerated artifacts: `using-ccode/accelerators.md`, `cookbook/accelerated-artifact.md`
- Project paths: `using-ccode/project-layout.md`, `reference/configuration.md`
- Minimal examples: `cookbook/minimal-process.md`

For OpenAPI schema or field-level details not covered by local docs, use the official external reference: `https://github.com/oai/openapi-specification`.

## Authoring Pattern

1. Read `ccode.yaml` before deciding where files belong.
2. Inspect `<ccode_path>/.ccode/lib/context.ts` when available to confirm the current runtime API.
3. Keep source inputs under `ccode_path`: processes, templates, specs, seed data, and instructions.
4. Put transformation and normalization logic in TypeScript. Keep templates focused on presentation.
5. Use stable artifact IDs and scopes. Call `ctx.setScope(...)` when the process filename is not the right state namespace.
6. Use `ctx.generate(...)` for deterministic files that should be regenerated.
7. Use `ctx.accelerate(...)` for proposed files that should not be overwritten after humans or agents adjust them.
8. Attach instruction markdown to accelerated artifacts when downstream adjustment requires domain guidance.
9. Make instructions specific to the target artifact and describe expected edits, constraints, and verification.

## Design Rules

- Do not write generator support files by hand under `<ccode_path>/.ccode/lib/*`; refresh them with `ccode init` when needed.
- Do not place instruction paths outside `ccode_path`.
- Do not rely on large branching logic inside templates when TypeScript can prepare explicit data.
- Do not create an accelerator for a file that should always be overwritten by the generator.
- Prefer small process/helper modules over one large process file when the workflow grows.

## Validation

When practical, run the exact process you changed:

```bash
ccode run <process>
```

Then inspect output files and accelerator state:

```bash
ccode list accelerated --for-agent
ccode list instructions --for-agent
```
