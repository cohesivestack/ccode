---
name: cohesive-code
description: Use when working with Cohesive Code or ccode projects, including initializing workspaces, authoring TypeScript processes, rendering templates, generating and accelerating artifacts, and running adjustment workflows from the CLI.
metadata:
  author: cohesivestack
  version: "1.1.0"
---

# Cohesive Code

Use this skill when the task involves `ccode`, `ccode.yaml`, `ccode init`, `ccode run`, TypeScript process files, Gonja templates, OpenAPI-driven generation, or accelerator adjustment workflows.

## Core rules

- Treat `ccode.yaml` as the source of truth for `ccode_path`, `output_path`, `hidden_path`, and required `version`.
- Keep authored sources under `ccode_path`. Do not hand-edit `.ccode/build/*`; it is generated cache.
- Avoid editing `.ccode/lib/*` in an application project; `ccode init` refreshes those generated support files.
- A runnable process must be a `.ts` file under `ccode_path` that exports `default function <name>(ctx: Context)`.
- Invoke processes as `ccode run <relative/process/path-without-.ts>`.
- `renderTemplate`, `generate`, `parseJSONFromFile`, and `parseOpenAPIFromFile` resolve template/input paths relative to `ccode_path`.
- `generate` writes relative outputs under `output_path`.
- `accelerate` writes artifacts under `output_path/<artifact-id>` and tracks them by scope in `.ccode/accelerators/<scope>/<artifact-id>.accelerated.json`.
- By default, accelerator scope is the process file name (without `.ts`); `ctx.setScope(...)` overrides it and `ctx.scope()` reads it.
- Prefer the generated `.ccode/lib/context.ts` contract over stale docs if README text and runtime behavior disagree.

## Workflow

1. Detect the workspace. Look for `ccode.yaml`. If the user is starting new or support files are missing/stale, initialize with `ccode init [path]` and optionally `--version <version>`.
2. Read `ccode.yaml` before editing files so path resolution is explicit.
3. Author or update the target process in TypeScript and import `type { Context } from "@ccode/context"`.
4. Keep templates, specs, and seed data inside `ccode_path` so the runner can resolve them consistently.
5. Validate with `ccode run <process>` and inspect output under `output_path`.
6. For accelerator workflows, use CLI metadata/instruction commands to review pending adjustments before editing target files.
7. Fix runtime errors from real command output rather than guessing.

## Authoring guidance

- Keep transformation logic in TypeScript. Use templates for presentation, not deep branching or normalization.
- Prefer small helper modules imported by the process over one large process file.
- Use `ctx.println` for temporary trace output while iterating.
- Use `ctx.parseJSONFromFile` when deterministic key order matters.
- Use `ctx.parseOpenAPIFromFile` only for OpenAPI v3.x input. Swagger/OpenAPI v2 is rejected.
- Pass plain objects and arrays into templates. Do not rely on class instances.
- Use `ctx.generate(...)` for standard artifact writes.
- Use `ctx.accelerate(...)` for artifacts that need human/agent adjustment and should avoid unsafe overwrite.

## Accelerator workflow guidance

- List pending adjustments: `ccode list accelerated [scopeId]`
- Inspect accelerator metadata: `ccode get accelerated <scopeId>:<artifactId>`
- Fetch adjustment bundle (instructions + decoded accelerated content): `ccode get accelerated <scopeId>:<artifactId> --instructions`
- List instruction references: `ccode list instructions`
- Read raw instruction file: `ccode get instruction <path>`
- For machine workflows, add `--for-agent` to list/get commands.

## Validation

- Prefer executing the exact process you changed.
- If `ccode` binary is unavailable and you are in the cohesive-code source repo, use `go run .`.
- Treat README as background context, not final API contract. Validate against generated context types and current CLI implementation.

## Read next as needed

- Path rules, config precedence, and initialized layout: [references/project-layout.md](references/project-layout.md)
- Process contract and runtime API: [references/runtime-api.md](references/runtime-api.md)
- OpenAPI generation patterns and pitfalls: [references/openapi-patterns.md](references/openapi-patterns.md)
- Accelerator adjustment flow details: [references/accelerator-workflow.md](references/accelerator-workflow.md)
- Concrete authoring examples: [examples/minimal-process.md](examples/minimal-process.md), [examples/openapi-generator.md](examples/openapi-generator.md), and [examples/accelerator-process.md](examples/accelerator-process.md)
- Shell helpers for inspection and smoke validation: [scripts/check_project.sh](scripts/check_project.sh) and [scripts/smoke_run.sh](scripts/smoke_run.sh)
