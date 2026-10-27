---
title: OpenAPI Workflows
description: Parse OpenAPI documents and generate deterministic artifacts.
---

Cohesive Code loads OpenAPI 3.0, 3.1, and 3.2 documents from bytes, strings, or files into versioned ergonomic models. File paths are relative to `ccode_path`. Local references retain their original declarations and source locations.

## OpenAPI reference

OpenAPI describes HTTP APIs with a language-agnostic YAML or JSON document that can drive documentation, client/server generation, and tests. Cohesive Code expects OpenAPI v3 input. When you need schema or field-level details, use the official [OAI/OpenAPI Specification repository](https://github.com/oai/openapi-specification) as the external reference.

## Recommended layout

```text
ccode/
  openapi/
    generate.proc.ts
  specs/
    api.yaml
  templates/
    docs/
      operations.tpl
```

## Recommended process shape

Load the spec once, select the fields needed by the template, then generate outputs.

```ts
import type { Context } from "@ccode/context";
import * as OpenAPI from "@ccode/openapi";

type OperationView = {
  method: string;
  path: string;
  operationId: string;
};

export default function main(ctx: Context) {
  const api = ctx.loadOpenAPIFromFile("specs/api.yaml", {
    expectedVersion: "3.1",
  });
  const operations: OperationView[] = api.operations.map(operation => ({
    method: operation.method.toUpperCase(),
    path: operation.path,
    operationId: operation.operationId ?? fallbackOperationId(operation.method, operation.path),
  }));

  for (const path of api.paths) {
    if (path.reference) {
      ctx.println(`loaded ${path.reference.targetSource.documentUri}`);
    }
  }

  ctx.generate("templates/docs/operations.tpl", "generated/operations.gen.md", {
    title: api.info.title,
    operations,
  });
}

function fallbackOperationId(method: string, path: string): string {
  return `${method}_${path.replace(/[^a-zA-Z0-9]+/g, "_")}`.replace(/^_+|_+$/g, "");
}
```

The model offers `api.operations`, effective `operation.parameters`, and scoped
`server.paths` / `server.operations`. Navigation shares object instances. Schema
references expose `schema.reference.target`; local constraints and composition
remain explicit. `api.raw` retains the original parsed root declaration.

For specification-shaped data, use `ctx.parseOpenAPIFromFile` and import types
from `@ccode/openapi/raw`. The existing file parser still materializes references
while preserving `$ref` fields. See [Runtime API](/reference/runtime-api/) for
both loading contracts and migration details.

Model elements expose reference metadata through `element.reference`, including
its original `uri`, `resolvedUri`, `source`, and `targetSource`. Schema references
also expose their resolved `target`.

Import `Path` from `@ccode/path` when a generated framework expects a
different path parameter syntax:

```ts
import * as Path from "@ccode/path";

Path.toColon("/users/{userId}"); // "/users/:userId"
Path.toSquareBrackets("/users/{userId}"); // "/users/[userId]"
Path.toCurlyBrackets("/users/:userId"); // "/users/{userId}"
Path.toCurlyBrackets("/users/[userId]"); // "/users/{userId}"
```

The helpers can also produce angle-bracket and dollar forms, and accept
`{ omitLeadingSlash: true }`. Each helper accepts all five supported syntaxes,
including mixed syntax. They do not substitute parameter values or validate
declarations against a Path Item or Operation. See [Runtime API](/reference/runtime-api/)
for parameter name rules and syntax limitations.

## Pitfalls

- Swagger/OpenAPI v2 is rejected.
- Pass `expectedVersion` when the process requires a specific OpenAPI version; parsing fails if the document declares a different version.
- `parseOpenAPIFromFile` resolves internal and external file references automatically. A missing file or fragment fails parsing with reference context rather than returning a partial document.
- Ergonomic collections default to empty arrays; optional element lookups still need guards.
- Model graphs can be passed directly to templates for navigation and method calls. Project the required fields before serialization or recursive filtering.
- Remote references are not fetched by the ergonomic loader.
- Some OpenAPI values can be references or booleans depending on location.
- Large templates that walk raw specs become hard to test and resolve.
- Validate the generated output, not only the process source.

## Generate or accelerate

Use `generate` for deterministic docs, indexes, schemas, and fully owned outputs.

Use `accelerate` for handlers, SDK files, examples, or implementation files that should receive edits after the generator proposes a starting point.
