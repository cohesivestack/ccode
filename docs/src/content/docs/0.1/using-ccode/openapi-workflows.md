---
title: OpenAPI Workflows
description: Parse OpenAPI documents and generate deterministic artifacts.
slug: 0.1/using-ccode/openapi-workflows
---

Cohesive Code can parse OpenAPI v3 documents from bytes, strings, or files. File parsing resolves paths relative to `ccode_path`, and local references are resolved from the spec directory.

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
  const spec = ctx.parseOpenAPIFromFile("specs/api.yaml");
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

## Pitfalls

* Swagger/OpenAPI v2 is rejected.
* Optional fields such as `paths`, `components`, and `schemas` need guards.
* Some OpenAPI values can be references or booleans depending on location.
* Large templates that walk raw specs become hard to test and adjust.
* Validate the generated output, not only the process source.

## Generate or accelerate

Use `generate` for deterministic docs, indexes, schemas, and fully owned outputs.

Use `accelerate` for handlers, SDK files, examples, or implementation files that should receive edits after the generator proposes a starting point.
