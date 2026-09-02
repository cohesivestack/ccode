---
title: Accelerated Artifact
description: Create a generated proposal with attached adjustment instructions.
---

## Layout

```text
ccode.yaml
ccode/
  api/
    generate.ts
  instructions/
    handlers.md
  templates/
    handlers.tpl
```

## Process

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  const model = {
    packageName: "handlers",
  };

  ctx.setScope("generate-api");

  ctx.accelerate(
    "handlers.go",
    "templates/handlers.tpl",
    model,
    "instructions/handlers.md",
  );
}
```

## Template

```jinja
package {{ data.packageName }}
```

## Instructions

```markdown
# Handler adjustment

Review the generated handler file, keep package naming stable, and add endpoint-specific behavior from the OpenAPI operation model.
```

## Run and inspect

```bash
ccode run api/generate
ccode list accelerated --for-agent
ccode get accelerated generate-api:handlers.go --instructions --for-agent
ccode adjust generate-api:handlers.go
```

The target file is `output_path/handlers.go`. If it is edited after generation, the next run will not overwrite it unless the file still matches the stored generated snapshot. Use `ccode adjust` when the proposal requires no edits and should be accepted as-is.
