# Runtime API

## Process contract

A runnable process is a TypeScript file under `ccode_path`:

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  ctx.println("hello");
}
```

Runner requirements:

- file ends in `.ts`
- default export is a function
- function takes one `Context` parameter
- CLI process argument omits `.ts`

If this signature does not match, the runner fails before execution.

## Import path

`ccode init` writes `tsconfig.json` so this resolves:

```ts
import type { Context } from "@ccode/context";
```

The alias points to `.ccode/lib/context.ts`.

## Current `Context` surface

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

Use generated `context.ts` in the workspace as local truth.

## Behavior notes

### `println`

- Writes to process stdout.

### `setScope` and `scope`

- Default scope is the process file name without `.ts`.
- `setScope("my-scope")` overrides the active scope.
- `scope()` returns the current scope.

### `renderTemplate(templatePath, data)`

- Loads Gonja template relative to `ccode_path`.
- Returns rendered string.

### `generate(templatePath, filePath, data)`

- Renders template relative to `ccode_path`.
- Writes to `filePath`.
- Relative output paths resolve under `output_path`.
- Parent directories are created automatically.

### `accelerate(id, templatePath, data, instructionsPath?)`

- Renders template relative to `ccode_path`.
- Targets file `output_path/<scope>/<id>`.
- Stores accelerator state at `.ccode/state/accelerators.json`.
- `instructionsPath` (optional) is stored as relative path under `ccode_path`.
- Performs safe-write behavior to avoid unsafe overwrite.
- Resets `adjusted_at` to `null` when a new accelerated version is written.

### `parseJSONFrom*`

- Returns a JS object.
- JSON root must be an object.
- `parseJSONFromFile` resolves paths relative to `ccode_path`.
- Preserves object key order.

### `parseOpenAPIFrom*`

- Returns a JSON-like JS object.
- `parseOpenAPIFromFile` resolves paths relative to `ccode_path`.
- `$ref` file resolution is based on spec directory.
- OpenAPI v3 only; Swagger/OpenAPI v2 rejected.
- Preserves object key order.

## Accelerator query CLI

These commands expose accelerator metadata and instructions:

- `ccode list accelerated [scopeId]`
- `ccode list instructions`
- `ccode get accelerated <scopeId>:<artifactId>`
- `ccode get accelerated <scopeId>:<artifactId> --instructions`
- `ccode get instruction <path>`

For machine-readable output, add `--for-agent`.

## Error handling

Go errors surface in TS as `GoError` and can be caught:

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

Common failures:

- missing template/input files
- invalid JSON/OpenAPI input
- unsupported Swagger/OpenAPI v2 input
- invalid process signature
