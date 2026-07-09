---
title: Generation
description: When to use ctx.generate and how generated files are written.
slug: 0.1/using-ccode/generation
---

Use `ctx.generate(...)` when the output is fully owned by the generator and should be overwritten on each run.

## API

```ts
ctx.generate(templatePath, filePath, data);
```

The runtime:

1. Resolves `templatePath` relative to `ccode_path`.
2. Renders the template with `data`.
3. Resolves relative `filePath` under `output_path`.
4. Creates parent directories.
5. Writes the rendered content.

## Example

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  ctx.generate("templates/status.tpl", "generated/status.md", {
    status: "alpha",
  });
}
```

Template:

```jinja
# Project status

{{ data.status }}
```

Result:

```text
generated/status.md
```

## When generate is the right choice

Use `generate` for:

* docs extracted from source metadata
* files that should match a schema exactly
* generated indexes
* temporary artifacts
* code that should never be hand-edited

Do not use `generate` for files that are expected to receive manual changes. Use [Accelerators](/0.1/using-ccode/accelerators/) instead.
