---
title: OpenAPI Docs
description: Generate a Markdown operation index from an OpenAPI document.
---

## Layout

```text
ccode.yaml
ccode/
  openapi/
    generate.proc.ts
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
  const api = ctx.loadOpenAPIFromFile("specs/api.yaml");
  const operations: OperationView[] = api.operations.map(operation => ({
    method: operation.method.toUpperCase(),
    path: operation.path,
    operationId: operation.operationId ?? buildFallbackId(operation.method, operation.path),
  }));

  ctx.generate("templates/docs/operations.tpl", "generated/operations.gen.md", {
    title: api.info.title,
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
