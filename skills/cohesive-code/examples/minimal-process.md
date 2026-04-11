# Minimal process example

Example layout:

```text
ccode.yaml
ccode/
  hello/
    generate.ts
  templates/
    greeting.tpl
```

`ccode/hello/generate.ts`

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

`ccode/templates/greeting.tpl`

```jinja
Hello {{ data.name }}!
```

Run it with:

```bash
ccode run hello/generate
```

Expected result:

- stdout prints `Hello Cohesive Code!`
- `output_path/generated/greeting.txt` is created

Why this pattern is good:

- the process contains only orchestration
- the template contains only formatting
- the generated file path is explicit
