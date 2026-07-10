---
title: AI Skill Index
description: A structured index for agents and future Cohesive Code skill documentation.
---

This page is written as an index for a future AI skill. It is not the skill itself. Use it as the stable map of concepts, files, commands, and docs pages an agent should read before acting in a Cohesive Code workspace.

## Detection

A Cohesive Code workspace usually has:

- `ccode.yaml` at or near the project root
- a `ccode_path` directory, defaulting to `ccode`
- generated support files under `<ccode_path>/.ccode/lib/`
- optional accelerator state under `<hidden_path>/accelerators/`

If `ccode.yaml` exists, read it before editing files.

## Source of truth order

Use this order when references disagree:

1. The generated local `Context` type at `<ccode_path>/.ccode/lib/context.ts`
2. The current CLI implementation and command output
3. This documentation
4. Older README examples

## Required reads by task

For initialization:

- [Getting Started](/getting-started/)
- [Project Layout](/using-ccode/project-layout/)
- [Configuration](/reference/configuration/)
- [Wrapper & Versions](/using-ccode/wrapper-and-versions/)

For authoring a process:

- [Processes](/using-ccode/processes/)
- [Templates](/using-ccode/templates/)
- [Runtime API](/reference/runtime-api/)
- [Minimal Process](/cookbook/minimal-process/)

For OpenAPI generation:

- [OpenAPI Workflows](/using-ccode/openapi-workflows/)
- [OpenAPI Docs](/cookbook/openapi-docs/)
- External schema reference: [OAI/OpenAPI Specification](https://github.com/oai/openapi-specification)
- [Runtime API](/reference/runtime-api/)

For accelerated artifacts:

- [Accelerators](/using-ccode/accelerators/)
- [Accelerator States](/reference/accelerator-states/)
- [Accelerated Artifact](/cookbook/accelerated-artifact/)
- [CLI](/reference/cli/)

## Agent operating rules

- Treat `ccode.yaml` as the workspace contract.
- Keep source files, templates, specs, and instructions under `ccode_path`.
- Do not hand-edit `.ccode/build/`.
- Do not edit `.ccode/lib/` in an application workspace; rerun `ccode init` to refresh it.
- Run the exact process you changed when validation is possible.
- Use `ctx.generate(...)` only for files that should be regenerated.
- Use `ctx.accelerate(...)` for files that may need human or agent adjustment.
- Inspect unresolved accelerator work with `ccode list accelerated --for-agent`.
- Fetch one adjustment bundle with `ccode get accelerated <scopeId>:<artifactId> --instructions --for-agent`.

## Command index

```bash
ccode init [path] --version <version>
ccode run <process>
ccode list accelerated [scopeId]
ccode list accelerated [scopeId] --include-resolved
ccode list instructions
ccode list instructions --include-resolved
ccode get accelerated <scopeId>:<artifactId>
ccode get accelerated <scopeId>:<artifactId> --instructions
ccode get instruction <path>
```

Add `--for-agent` to list and get commands when the caller needs machine-readable JSON.

## Runtime API index

```ts
ctx.println(message)
ctx.setScope(scopeName)
ctx.scope()
ctx.renderTemplate(templatePath, data)
ctx.generate(templatePath, filePath, data)
ctx.accelerate(id, templatePath, data, instructionsPath?)
ctx.parseJSONFromBytes(jsonBytes)
ctx.parseJSONFromString(jsonString)
ctx.parseJSONFromFile(filePath)
ctx.parseOpenAPIFromBytes(specBytes)
ctx.parseOpenAPIFromString(spec)
ctx.parseOpenAPIFromFile(filePath)
```
