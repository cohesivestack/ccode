# Accelerator process example

Example layout:

```text
ccode.yaml
ccode/
  api/
    generate.ts
  templates/
    handlers.tpl
  instructions/
    handlers.md
```

`ccode/api/generate.ts`

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  const model = {
    packageName: "handlers",
  };

  // Optional: group accelerated artifacts under a custom scope.
  ctx.setScope("generate-api");

  ctx.accelerate(
    "handlers.go",
    "templates/handlers.tpl",
    model,
    "instructions/handlers.md",
  );
}
```

`ccode/templates/handlers.tpl`

```jinja
package {{ data.packageName }}
```

Adjustment flow:

```bash
ccode run api/generate
ccode list accelerated
ccode get accelerated generate-api:handlers.go --instructions
```

Why this pattern is good:

- accelerator output is tracked in state
- instructions are attached to the artifact
- reviewers/agents can fetch a single adjustment bundle from CLI
