# OpenAPI patterns

## Recommended layout

Keep the spec, process, and templates together under `ccode_path`:

```text
ccode/
  openapi/
    generate.ts
  specs/
    api.yaml
  templates/
    docs/
      operations.tpl
    sdk/
      client.tpl
```

## Recommended workflow

1. Parse the spec once with `ctx.parseOpenAPIFromFile("specs/api.yaml")`.
2. Build a template-friendly model in TypeScript.
3. Keep guards for optional fields and union-like shapes.
4. Render one or more outputs with `generate` (or `accelerate` when review/adjustment is required).
5. Validate the generated artifacts, not just the process source.

## Process pattern

Use TypeScript to flatten the OpenAPI document into a smaller rendering model:

```ts
import type { Context } from "@ccode/context";

type OperationDoc = {
  method: string;
  path: string;
  operationId: string;
  summary: string;
};

export default function main(ctx: Context) {
  const spec = ctx.parseOpenAPIFromFile("specs/api.yaml");
  const operations: OperationDoc[] = [];

  for (const [path, item] of Object.entries(spec.paths ?? {})) {
    if (!item) continue;

    for (const method of ["get", "post", "put", "patch", "delete"] as const) {
      const operation = item[method];
      if (!operation) continue;

      operations.push({
        method: method.toUpperCase(),
        path,
        operationId: operation.operationId ?? fallbackOperationId(method, path),
        summary: operation.summary ?? "",
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

## Template pattern

Keep templates shallow. Example:

```jinja
# {{ data.title }}

{% for operation in data.operations %}
- `{{ operation.method }} {{ operation.path }}` {% if operation.summary %}- {{ operation.summary }}{% endif %}
{% endfor %}
```

## Pitfalls

- Do not pass the whole OpenAPI document directly into many large templates if you only need a few fields.
- Guard for optional `paths`, `components`, `schemas`, and operation fields.
- Schema values may be booleans or references in some locations. Check shape before reading nested properties.
- Reject or convert Swagger/OpenAPI v2 before calling the parser.
- Keep templates deterministic. Avoid making them responsible for schema walking or path normalization.

## Validation checklist

- The input file is OpenAPI v3.x.
- Local `$ref` and file references resolve correctly from the spec directory.
- The generated output lands under `output_path`.
- The process is re-runnable without manual cleanup.
- If using `accelerate`, unresolved items are visible in `ccode list accelerated`.
