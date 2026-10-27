---
title: Accelerated Artifact
description: Create a generated proposal with attached resolvement guides.
---

## Layout

```text
ccode.yaml
ccode/
  api/
    generate.proc.ts
  guides/
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

  ctx.accelerate(
    "handlers.acc.go",
    "templates/handlers.tpl",
    model,
    "guides/handlers.md",
  );
}
```

## Template

```jinja
package {{ data.packageName }}
```

## Guides

```markdown
# Handler resolvement

Review the generated handler file, keep package naming stable, and add endpoint-specific behavior from the OpenAPI operation model.
```

## Run and inspect

```bash
ccode run api/generate
ccode acc list --for-agent
ccode acc show handlers.acc.go --guide --for-agent
ccode acc resolve handlers.acc.go
```

The target file is `output_path/handlers.acc.go`. If it is edited after generation, the next run will not overwrite it unless the file still matches the stored generated snapshot. Use `ccode acc resolve` when the proposal requires no edits and should be accepted as-is.
