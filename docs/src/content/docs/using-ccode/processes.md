---
title: Processes
description: The TypeScript process contract and authoring style.
---

A process is a TypeScript file that Cohesive Code compiles and runs. It should orchestrate input parsing, model shaping, template rendering, generation, and acceleration.

## Contract

Every runnable process must be under `ccode_path`, end in `.process.ts`, and default-export a function. The runtime passes a `Context` value as the function's first argument. The function may be named or anonymous.

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  ctx.println("hello");
}
```

The filename is the explicit marker that distinguishes runnable processes from ordinary TypeScript helper modules. After compiling the process, the runner reads the module's default export, verifies that it is callable, and invokes it with the runtime context.

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

## Running another process

Call `runProcess` to run another process synchronously with the same resolved
configuration:

```ts
export default function main(ctx: Context) {
  ctx.runProcess("api/generate");
}
```

The child process gets its own JavaScript runtime and runner context. If it
fails, the error stops the calling process unless it is caught.
Circular executions, such as `a` running `b` and `b` running `a`, return an
error.

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
