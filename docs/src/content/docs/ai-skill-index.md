---
title: AI Skill Index
description: The documentation map used by the installable Cohesive Code agent skill.
---

This page is the stable documentation map for the installable `cohesive-code` skill. It indexes the concepts, files, commands, and pages an agent should read before acting in a Cohesive Code workspace; the skill itself lives at `skills/cohesive-code/SKILL.md` in the repository.

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
- Do not hand-edit `.ccode/build/`.
- Do not edit `.ccode/lib/` in an application workspace; rerun `ccode init` to refresh it.
- Run the exact process you changed when validation is possible.
- Use `ctx.generate(...)` only for files that should be regenerated.
- Use `ctx.accelerate(...)` for files that may need human or agent resolvement.
- Inspect unresolved accelerator work with `ccode acc list --for-agent`.
- Fetch one resolvement bundle with `ccode acc show <artifactId> --guide --for-agent`.

## Command index

```bash
ccode init [path] --version <version>
ccode run <process>
ccode acc resolve <artifactId>
ccode proc list
ccode acc list
ccode acc list --include-resolved
ccode acc show <artifactId>
ccode acc show <artifactId> --guide
ccode acc guide <artifactId>
```

`ccode proc list` always returns machine-readable JSON. Add `--for-agent` to accelerator list and show commands when the caller needs structured output.

## Runtime API index

```ts
ctx.println(message)
ctx.renderTemplate(templatePath, data)
ctx.generate(templatePath, filePath, data)
ctx.accelerate(id, templatePath, data, guidePath?)
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
