---
name: cohesive-code
description: Use when working with Cohesive Code or ccode projects, including initializing a workspace, editing ccode.yaml config, authoring or fixing TypeScript processes, rendering Gonja templates, using OpenAPI-driven generators, and validating ccode run workflows across editors and agents.
---

# Cohesive Code

Use this skill when the task involves `ccode`, `ccode.yaml`, `ccode init`, `ccode run`, TypeScript process files, Gonja templates, or OpenAPI-based generation.

## Core rules

- Treat `ccode.yaml` as the source of truth for `ccode_path`, `output_path`, and `hidden_path`. The `path` key is legacy and maps to `ccode_path`.
- Keep authored sources under `ccode_path`. Do not hand-edit `.ccode/build/*`; it is generated cache.
- Avoid editing `.ccode/lib/*` in an application project unless the task is to change the Cohesive Code CLI templates themselves.
- A runnable process must be a `.ts` file under `ccode_path` that exports `default function <name>(ctx: Context)`.
- Invoke processes as `ccode run <relative/process/path-without-.ts>`.
- `templateToString`, `templateToFile`, `parseJSONFromFile`, and `parseOpenAPIFromFile` resolve input paths relative to `ccode_path`.
- `templateToFile` writes relative output paths under `output_path`.
- Prefer the generated `.ccode/lib/context.ts` contract over stale docs if the README and runtime disagree.

## Workflow

1. Detect the workspace. Look for `ccode.yaml`. If the user is starting a new generator, initialize one with `ccode init [path]`.
2. Read `ccode.yaml` before editing files so path resolution is explicit.
3. Author or update the target process in TypeScript and import `type { Context } from "@ccode/context"`.
4. Keep templates, specs, and seed data inside `ccode_path` so the runner can resolve them consistently.
5. Validate with `ccode run <process>` and inspect the generated output under `output_path`.
6. Fix runtime errors from the actual runner output rather than guessing. Common failures are wrong process signatures, missing files, bad relative paths, JSON roots that are not objects, and Swagger/OpenAPI v2 input.

## Authoring guidance

- Keep transformation logic in TypeScript. Use templates for presentation, not for deep branching or normalization.
- Prefer small helper modules imported by the process over one large process file.
- Use `ctx.println` for temporary trace output while iterating.
- Use `ctx.parseJSONFromFile` when you need deterministic key order from source JSON.
- Use `ctx.parseOpenAPIFromFile` only for OpenAPI v3.x input. The current runtime rejects Swagger/OpenAPI v2.
- Pass plain objects and arrays into templates. Do not rely on JS class instances.
- Do not assume undocumented helpers exist. The current runtime contract is limited and explicit.

## Validation

- Prefer executing the exact process you changed.
- If the `ccode` binary is not installed but you are working inside the cohesive-code source repo, use `go run .`.
- Treat the README as background context, not the final API contract. Validate against the generated context types or the current CLI implementation.

## Read next as needed

- Path rules, config precedence, and initialized layout: [references/project-layout.md](references/project-layout.md)
- Process contract and runtime API: [references/runtime-api.md](references/runtime-api.md)
- OpenAPI generation patterns and pitfalls: [references/openapi-patterns.md](references/openapi-patterns.md)
- Concrete authoring examples: [examples/minimal-process.md](examples/minimal-process.md) and [examples/openapi-generator.md](examples/openapi-generator.md)
- Shell helpers for inspection and smoke validation: [scripts/check_project.sh](scripts/check_project.sh) and [scripts/smoke_run.sh](scripts/smoke_run.sh)
