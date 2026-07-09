---
title: OpenAPI Docs
description: Generate a Markdown operation index from an OpenAPI document.
slug: 0.1/cookbook/openapi-docs
---

## Layout

```text
ccode.yaml
ccode/
  openapi/
    generate.ts
  specs/
    api.yaml
  templates/
    docs/
      operations.tpl
```

## Process

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
        operationId: operation.operationId ?? buildFallbackId(method, path),
      });
    }
  }

  ctx.generate("templates/docs/operations.tpl", "generated/operations.md", {
    title: spec.info?.title ?? "API",
    operations,
  });
}

function buildFallbackId(method: string, path: string): string {
  return `${method}_${path.replace(/[^a-zA-Z0-9]+/g, "_")}`.replace(/^_+|_+$/g, "");
}
```

## Template

```jinja
# {{ data.title }}

{% for operation in data.operations %}
- `{{ operation.operationId }}` -> `{{ operation.method }} {{ operation.path }}`
{% endfor %}
```

## Run

```bash
ccode run openapi/generate
```

This pattern keeps OpenAPI traversal in TypeScript and leaves Markdown formatting to the template.
