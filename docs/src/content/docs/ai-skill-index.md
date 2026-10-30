---
title: AI Skill Index
description: The documentation map used by the installable Cohesive Code agent skills.
---

This page is the stable documentation map for the installable `ccode-author` and `ccode-produce` skills. It indexes the concepts, files, commands, and pages an agent should read before acting in a Cohesive Code workspace. The skills live at `skills/ccode-author/SKILL.md` and `skills/ccode-produce/SKILL.md` in the repository.

- Use `ccode-author` to initialize projects and create or modify processes, templates, inputs, and accelerator guides.
- Use `ccode-produce` to run existing processes, apply guided accelerated changes, validate, and resolve artifacts.

## Detection

A Cohesive Code workspace usually has:

- `ccode.yaml` at or near the project root
- a `ccode_path` directory, defaulting to `ccode`
- generated support files under `<hidden_path>/lib/` (default: `<ccode_path>/.ccode/lib/`)
- optional accelerator state under `<hidden_path>/accelerators/`

If `ccode.yaml` exists, read it before editing files.

## Source of truth order

Use this order when references disagree:

1. The generated local `Context` type at `<hidden_path>/lib/context.ts`
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

For database schema generation:

- [Database Workflows](/using-ccode/database-workflows/)
- [Database Inspection](/reference/database-inspection/)
- [Database Schema](/cookbook/database-schema/)
- [Runtime API](/reference/runtime-api/)

For accelerated artifacts:

- [Accelerators](/using-ccode/accelerators/)
- [Accelerator States](/reference/accelerator-states/)
- [Accelerated Artifact](/cookbook/accelerated-artifact/)
- [CLI](/reference/cli/)

## Agent operating rules

- Treat `ccode.yaml` as the workspace contract.
- Keep source files, templates, specs, and guides under `ccode_path`.
- Do not hand-edit `<hidden_path>/build/`.
- Do not edit `<hidden_path>/lib/` in an application workspace; use `ccode-author` to refresh it through `ccode init`.
- Use `ccode-produce` with the exact changed process when execution is needed for authoring validation.
- Use `ctx.generate(...)` only for files that should be regenerated.
- Use `ctx.accelerate(...)` for files that may need human or agent resolvement.
- Inspect unresolved accelerator work with `ccode acc list --json`.
- Request individual artifact metadata with `ccode acc show <artifactId> --json`; the default show output is YAML.
- Read each pending artifact's guide with `ccode acc guide <artifactId>`, compare its saved snapshot to the current output with `ccode acc diff <artifactId>`, and retrieve the snapshot with `ccode acc content <artifactId>`.
- Leave artifacts without guides pending and report them. Resolve completed artifacts with `ccode acc resolve <artifactId>` only after validation succeeds.

## Command index

```bash
ccode init [path] --version <version>
ccode run <process>
ccode acc resolve <artifactId>
ccode proc list
ccode proc add <process>
ccode acc list
ccode acc list --json
ccode acc list --include-resolved
ccode acc show <artifactId>
ccode acc show <artifactId> --json
ccode acc content <artifactId>
ccode acc diff <artifactId>
ccode acc guide <artifactId>
```

`ccode proc list` always returns machine-readable JSON. `ccode acc list` and
`ccode acc show` return YAML metadata by default; use `--json` explicitly when
parsing their output. `ccode acc guide` returns only raw Markdown without a path
or JSON wrapper. Retrieve snapshots and diffs with the separate commands above.

## Runtime API index

```ts
ctx.println(message)
ctx.renderTemplate(templatePath, data)
ctx.generate(templatePath, filePath, data)
ctx.accelerate(id, templatePath, data, guidePath?)
ctx.prune(pathPatterns)
ctx.parseJSONFromBytes(jsonBytes)
ctx.parseJSONFromString(jsonString)
ctx.parseJSONFromFile(filePath)
ctx.parseOpenAPIFromBytes(specBytes)
ctx.parseOpenAPIFromString(spec)
ctx.parseOpenAPIFromFile(filePath)
ctx.inspectDatabase(connectionURL)
ctx.inspectDatabase(connectionURL, { expectedEngine })
Path.toCurlyBrackets(path, options?)
Path.toColon(path, options?)
Path.toSquareBrackets(path, options?)
Path.toAngleBrackets(path, options?)
Path.toDollar(path, options?)
Strings.camelCase(value, initialisms?)
Strings.pascalCase(value, initialisms?)
Strings.snakeCase(value)
Strings.kebabCase(value)
Strings.constantCase(value)
Strings.dotCase(value)
Strings.pathCase(value)
Strings.titleCase(value, initialisms?)
Strings.sentenceCase(value, initialisms?)
Strings.upperFirst(value)
Strings.lowerFirst(value)
Strings.normalizeSpace(value)
Go.toExportedIdentifier(value, initialisms?)
Go.toUnexportedIdentifier(value, initialisms?)
Go.toPackageName(value)
Go.toType(value, options?)
TypeScript.toType(value, options?)
TypeScript.toTypeIdentifier(value, initialisms?)
TypeScript.toValueIdentifier(value, initialisms?)
```

Import `OpenAPI` from `@ccode/openapi`, `Path` from `@ccode/path`, `Strings`
from `@ccode/string`, `Go` from `@ccode/go`, and `TypeScript` from
`@ccode/typescript`. Templates provide equivalent string, Go naming,
TypeScript naming, Go primitive type inference (`goType`), and route path
filters; see [Templates](/using-ccode/templates/).
