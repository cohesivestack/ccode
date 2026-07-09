---
title: Runtime API
description: The TypeScript Context API exposed to Cohesive Code processes.
---

Import `Context` from the generated workspace support files:

```ts
import type { Context } from "@ccode/context";
```

## Surface

```ts
interface Context {
  println(message: string): void;
  setScope(scopeName: string): void;
  scope(): string;
  renderTemplate(templatePath: string, data: any): string;
  generate(templatePath: string, filePath: string, data: any): void;
  accelerate(
    id: string,
    templatePath: string,
    data: any,
    instructionsPath?: string,
  ): void;
  parseJSONFromBytes(jsonBytes: number[]): Record<string, any>;
  parseJSONFromString(jsonString: string): Record<string, any>;
  parseJSONFromFile(filePath: string): Record<string, any>;
  parseOpenAPIFromBytes(specBytes: number[]): OpenAPIDocument;
  parseOpenAPIFromString(spec: string): OpenAPIDocument;
  parseOpenAPIFromFile(filePath: string): OpenAPIDocument;
}
```

Use the local generated `context.ts` as the final contract for the installed workspace version.

## println

Writes to process stdout.

```ts
ctx.println("building docs");
```

## scope and setScope

Reads or overrides the active accelerator scope.

```ts
ctx.println(ctx.scope());
ctx.setScope("api-handlers");
```

## renderTemplate

Renders a Gonja template and returns a string.

```ts
const text = ctx.renderTemplate("templates/readme.tpl", model);
```

## generate

Renders a template and writes it to an output file.

```ts
ctx.generate("templates/readme.tpl", "README.generated.md", model);
```

## accelerate

Renders a template, tracks generated state, and writes safely to `output_path/<id>`.

```ts
ctx.accelerate("src/handlers.ts", "templates/handlers.tpl", model, "instructions/handlers.md");
```

## parseJSONFrom*

Parses JSON and returns a JavaScript object. The root JSON value must be an object.

```ts
const settings = ctx.parseJSONFromFile("data/settings.json");
```

## parseOpenAPIFrom*

Parses OpenAPI v3 input and returns a JSON-like object. `parseOpenAPIFromFile` resolves paths relative to `ccode_path`.

```ts
const spec = ctx.parseOpenAPIFromFile("specs/api.yaml");
```
