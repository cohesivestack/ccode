---
title: OpenAPI Workflows
description: Parse OpenAPI documents and generate deterministic artifacts.
---

Cohesive Code can parse OpenAPI v3 documents from bytes, strings, or files. File parsing resolves paths relative to `ccode_path`. Local and external file references are resolved recursively by default, relative to the file that contains each reference.

## OpenAPI reference

OpenAPI describes HTTP APIs with a language-agnostic YAML or JSON document that can drive documentation, client/server generation, and tests. Cohesive Code expects OpenAPI v3 input. When you need schema or field-level details, use the official [OAI/OpenAPI Specification repository](https://github.com/oai/openapi-specification) as the external reference.

## Recommended layout

```text
ccode/
  openapi/
    generate.ts
  specs/
    api.yaml
  templates/
    docs/
      operations.tpl
```

## Recommended process shape

Parse the spec once, reduce it to a template-friendly model, then generate outputs.

```ts
import type { Context } from "@ccode/context";

type OperationView = {
  method: string;
  path: string;
  operationId: string;
};

export default function main(ctx: Context) {
  const spec = ctx.parseOpenAPIFromFile("specs/api.yaml", {
    expectedVersion: "3.1",
  });
  const operations: OperationView[] = [];

  for (const [path, item] of Object.entries(spec.paths ?? {})) {
    if (!item) continue;

    for (const method of ["get", "post", "put", "patch", "delete"] as const) {
      const operation = item[method];
      if (!operation) continue;

      operations.push({
        method: method.toUpperCase(),
        path,
        operationId: operation.operationId ?? fallbackOperationId(method, path),
      });
    }
  }

  ctx.generate("templates/docs/operations.tpl", "generated/operations.md", {
    title: spec.info?.title ?? "API",
    operations,
  });
}

function fallbackOperationId(method: string, path: string): string {
  return `${method}_${path.replace(/[^a-zA-Z0-9]+/g, "_")}`.replace(/^_+|_+$/g, "");
}
```

Referenced Path Items are materialized without losing their source reference. For example, a Path Item declared as `$ref: ./paths/countries.yaml#/countries` exposes both `item.$ref` and resolved operations such as `item.get`. Generators can therefore infer the source file directly from `$ref`; no separate manifest or second parse is needed. Nested references, including schemas referenced from Path Item files, resolve automatically as well.

## Pitfalls

- Swagger/OpenAPI v2 is rejected.
- Pass `expectedVersion` when the process requires a specific OpenAPI version; parsing fails if the document declares a different version.
- `parseOpenAPIFromFile` resolves internal and external file references automatically. A missing file or fragment fails parsing with reference context rather than returning a partial document.
- Optional fields such as `paths`, `components`, and `schemas` need guards.
- Some OpenAPI values can be references or booleans depending on location.
- Large templates that walk raw specs become hard to test and adjust.
- Validate the generated output, not only the process source.

## Generate or accelerate

Use `generate` for deterministic docs, indexes, schemas, and fully owned outputs.

Use `accelerate` for handlers, SDK files, examples, or implementation files that should receive edits after the generator proposes a starting point.
