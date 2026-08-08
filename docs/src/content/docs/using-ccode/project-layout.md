---
title: Project Layout
description: Workspace files, path resolution, and recommended source organization.
---

Cohesive Code separates authored sources, generated outputs, and internal state. Keep that separation intact so processes stay rerunnable and agents can reason about the project safely.

## Default layout

Running `ccode init ccode --version v0.1.0` creates:

```text
ccode.yaml
ccode/
  tsconfig.json
  .ccode/
    .gitignore
    accelerators/
    build/
    lib/
      context.ts
      openapi.ts
      database.ts
      database-postgresql.ts
      database-mysql.ts
      database-mariadb.ts
      database-sqlite.ts
```

## Important folders

`ccode_path` is where authored generator assets live. The default is `ccode`.

`output_path` is where generated artifacts are written. The default is `.`.

`hidden_path` is internal state. The default is `.ccode`, and relative values resolve under `ccode_path`.

`<hidden_path>/lib/` contains generated TypeScript support files. Do not edit these in application workspaces.

`<hidden_path>/build/` is compiler cache and can be recreated.

`<hidden_path>/accelerators/` stores accelerator state. It is generated state, but it is meaningful and may be committed when a workflow needs persistent adjustment metadata.

## Recommended organization

The runtime does not require a deep structure. This layout keeps larger projects readable:

```text
ccode/
  api/
    generate.ts
  data/
    seed.json
  instructions/
    handlers.md
  specs/
    api.yaml
  templates/
    docs/
      operations.tpl
    sdk/
      client.tpl
```

Use stable folders for process entrypoints, templates, input specs, and adjustment instructions. Keep files inside `ccode_path` unless there is a clear reason to use external inputs.

## Process paths

`ccode run <process>` resolves `<process>` relative to `ccode_path`.

Valid:

```bash
ccode run api/generate
```

Invalid:

```bash
ccode run ../outside
ccode run /absolute/path
ccode run .
```

The CLI appends `.ts` when the process argument does not include it.

## Path resolution summary

- Template paths resolve relative to `ccode_path`.
- JSON file paths resolve relative to `ccode_path`.
- OpenAPI file paths resolve relative to `ccode_path`.
- Database connection URLs are passed directly to `inspectDatabase`; they do not resolve through `ccode_path`.
- Relative `generate` output paths resolve under `output_path`.
- `accelerate` targets `output_path/<artifact-id>`.
- Relative `hidden_path` resolves under `ccode_path`.
