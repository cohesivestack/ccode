---
title: Minimal Process
description: Render and write a simple generated text file.
---

## Layout

```text
ccode.yaml
ccode/
  hello/
    generate.ts
  templates/
    greeting.tpl
```

## Process

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  const model = {
    name: "Cohesive Code",
  };

  const preview = ctx.renderTemplate("templates/greeting.tpl", model);
  ctx.println(preview);
  ctx.generate("templates/greeting.tpl", "generated/greeting.txt", model);
}
```

## Template

```jinja
Hello {{ data.name }}!
```

## Run

```bash
ccode run hello/generate
```

Expected result:

- stdout prints `Hello Cohesive Code!`
- `output_path/generated/greeting.txt` is created

This pattern is useful because the process handles orchestration, the template handles formatting, and the output path is explicit.
