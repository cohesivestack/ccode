---
title: Processes
description: The TypeScript process contract and authoring style.
slug: 0.1/using-ccode/processes
---

A process is a TypeScript file that Cohesive Code compiles and runs. It should orchestrate input parsing, model shaping, template rendering, generation, and acceleration.

## Contract

Every runnable process must be under `ccode_path`, end in `.ts`, and export a default function with one `Context` parameter.

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  ctx.println("hello");
}
```

The runner validates this signature before execution. If the source does not match, the command fails before running the bundle.

## Import path

`ccode init` writes a `tsconfig.json` so this import resolves:

```ts
import type { Context } from "@ccode/context";
```

The alias points to the generated local contract at `.ccode/lib/context.ts`.

## Authoring style

Prefer this shape:

1. Parse external input once.
2. Normalize it into a small TypeScript model.
3. Pass plain objects and arrays to templates.
4. Generate or accelerate explicit outputs.
5. Print only useful trace information.

Example:

```ts
import type { Context } from "@ccode/context";

type Page = {
  title: string;
  slug: string;
};

export default function main(ctx: Context) {
  const data = ctx.parseJSONFromFile("data/pages.json");
  const pages: Page[] = data.pages.map((page: any) => ({
    title: page.title,
    slug: page.slug,
  }));

  ctx.generate("templates/pages.tpl", "generated/pages.md", { pages });
}
```

## Scope

Each process has an active accelerator scope. The default is the process file name without `.ts`.

```ts
ctx.scope();
ctx.setScope("openapi-handlers");
```

Use a custom scope when one process emits artifacts that should be grouped under a durable workflow name instead of the file name.

## Error handling

Runtime errors from Go functions surface inside JavaScript execution. Let failures stop the process unless the process has a meaningful recovery path.

```ts
try {
  ctx.parseOpenAPIFromFile("specs/api.yaml");
} catch (error) {
  ctx.println(String(error));
  return;
}
```

For generator workflows, failing fast is usually better than producing partial output from invalid input.
