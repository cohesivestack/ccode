# Runtime API

## Process contract

A runnable process is a TypeScript file under `ccode_path` that matches this shape:

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  ctx.println("hello");
}
```

Current runner requirements:

- The file must end in `.ts`.
- The process must export a default function.
- The default function must take a single parameter typed as `Context`.
- The CLI process argument omits the `.ts` suffix.

If the signature does not match, the runner fails before execution.

## Import path

`ccode init` writes `tsconfig.json` so this import works inside the project:

```ts
import type { Context } from "@ccode/context";
```

The alias resolves to `.ccode/lib/context.ts`.

## Current `Context` surface

The runtime currently exposes these methods:

```ts
interface Context {
  println(message: string): void;
  templateToString(templatePath: string, data: any): string;
  templateToFile(templatePath: string, filePath: string, data: any): void;
  parseJSONFromBytes(jsonBytes: number[]): Record<string, any>;
  parseJSONFromString(jsonString: string): Record<string, any>;
  parseJSONFromFile(filePath: string): Record<string, any>;
  parseOpenAPIFromBytes(specBytes: number[]): OpenAPIDocument;
  parseOpenAPIFromString(spec: string): OpenAPIDocument;
  parseOpenAPIFromFile(filePath: string): OpenAPIDocument;
}
```

Use the generated `context.ts` in the workspace as the local truth.

## Behavior notes

### `println`

- Use string output for portability.
- It writes to the process stdout stream.

### `templateToString(templatePath, data)`

- Loads a Gonja template relative to `ccode_path`.
- Returns the rendered string.
- Best for previews, testing, or composing a larger output in TypeScript.

### `templateToFile(templatePath, filePath, data)`

- Loads a Gonja template relative to `ccode_path`.
- Writes the rendered result to `filePath`.
- Relative output paths are resolved under `output_path`.
- Parent directories are created automatically.

Important:

- The current runtime signature is `(templatePath, filePath, data)`.
- Older README text may show a different parameter order or optional overwrite flags. Do not follow that older signature.

### `parseJSONFrom*`

- All three variants return a JS object.
- The JSON root must be an object. Arrays, numbers, strings, and `null` are rejected.
- `parseJSONFromFile` resolves the input path relative to `ccode_path`.
- Key order from the source object is preserved by the runtime.

### `parseOpenAPIFrom*`

- Parses OpenAPI input and returns a JSON-like JS object.
- `parseOpenAPIFromFile` resolves the input path relative to `ccode_path`.
- File references are resolved from the OpenAPI file directory.
- Only OpenAPI v3 is supported.
- Swagger/OpenAPI v2 input is rejected.
- Key order from the parsed document is preserved.

## Error handling

Errors raised from Go surface in TypeScript as `GoError` values and can be caught:

```ts
try {
  ctx.parseOpenAPIFromFile("specs/missing.yaml");
} catch (e: any) {
  if (e instanceof GoError) {
    ctx.println(e.value.Error());
    return;
  }
  throw e;
}
```

Common runtime failures:

- missing template file
- missing JSON or spec file
- invalid JSON input
- invalid OpenAPI input
- unsupported Swagger/OpenAPI v2 input
- process signature not matching `default function <name>(ctx: Context)`

## Guidance for agents

- Normalize data in TypeScript before rendering templates.
- Keep templates declarative and readable.
- Trust the current generated API types over aspirational documentation.
