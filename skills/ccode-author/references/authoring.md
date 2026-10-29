# Authoring contracts

Use installed public declarations for exact signatures. Source precedence is the
project's generated API, matching CLI implementation/output, matching docs, then
older examples. Report version mismatches rather than silently using stale APIs.

## Processes and templates

- Runnable entry points end in `.proc.ts`, default-export a function, and use
  `import type { Context } from "@ccode/context"`. Other TypeScript files are helpers.
- Parse each input once, shape template data in TypeScript, and keep Gonja
  templates focused on presentation. The root template variable is `data`.
- `ctx.generate(templatePath, outputPath, data)` owns a file the generator may
  overwrite. `ctx.accelerate(id, templatePath, data, guidePath?)` proposes a file
  whose later human or agent edits must be preserved. Note the different argument
  order. Templates and guides resolve under `ccode_path`; outputs under `output_path`.
- `ctx.runProcess("api/generate")` invokes a child synchronously with the same
  configuration. Avoid cycles. Inspect pruning semantics before adding destructive
  cleanup; place pruning after the relevant processes have emitted their outputs.
- Guides explain the intended adaptation, constraints, compatible edits to
  preserve, and concrete validation. Use stable IDs matching relative output paths.

For example, after scaffolding a process and authoring the referenced files:

```ts
import type { Context } from "@ccode/context";

export default function run(ctx: Context) {
  const data = ctx.parseJSONFromFile("data/service.json");
  ctx.generate("templates/model.tpl", "src/model.gen.ts", data);
  ctx.accelerate(
    "src/handler.ts", "templates/handler.tpl", data, "guides/handler.md",
  );
}
```

```jinja
export interface {{ data.name | typeScriptType }} {
  id: string;
}
```

## Reuse public helpers

| Need | Module and contract to inspect |
| --- | --- |
| General case and whitespace | `@ccode/string`: case functions and `normalizeSpace` |
| Go declarations and primitive types | `@ccode/go`: exported/unexported identifiers, package names, `toType` |
| TypeScript declarations and primitive types | `@ccode/typescript`: `toTypeIdentifier`, `toValueIdentifier`, `toType` |
| Route parameter syntax | `@ccode/path`: `toCurlyBrackets`, `toColon`, `toSquareBrackets`, `toAngleBrackets`, `toDollar` |
| OpenAPI models and references | `@ccode/openapi`, `@ccode/openapi/raw`, and the installed `Context` overloads |
| Database inspection models and guards | `@ccode/database` and the installed `Context` overloads |

Do not reimplement available naming/type/path helpers. Go helpers include
conventional initialisms; generic and TypeScript naming need an explicit list
when spellings such as `API` or `ID` matter. Type inference handles primitives;
keep references, containers, nullability, imports, and serialization decisions in
the generator. Unknown types fail by default; choose explicit overrides/fallbacks
only when project requirements justify them.

Gonja exposes corresponding naming and route filters; verify supported names and
keyword arguments in matching docs. TypeScript naming filters are `typeScriptType`
and `typeScriptValue`; these normalize identifiers, not primitive type expressions.
Route helpers convert syntax without interpolating or validating declarations.

For OpenAPI, distinguish raw parsed documents from ergonomic loaded models. Use
`import type` for model declarations. For ergonomic references use
`element.reference` for source metadata and narrow `schema.kind` before following
`schema.reference.target`. For template traversal, follow the matching version's
model and renderer documentation; do not assume all objects must be plain data.

## Read only the documentation needed

In a ccode checkout, use `docs/src/content/docs/`. Otherwise retrieve the matching
page at `https://cohesivecode.dev/` or its source under
`https://raw.githubusercontent.com/cohesivestack/ccode/master/docs/src/content/docs/`.
Current online docs can differ from the project's pinned version.

| Task | Documentation paths |
| --- | --- |
| Initialization/configuration | `getting-started.md`, `reference/configuration.md`, `using-ccode/project-layout.md`, `using-ccode/wrapper-and-versions.md` |
| Processes/templates | `using-ccode/processes.md`, `using-ccode/templates.md`, `reference/runtime-api.md` |
| Accelerators/guides | `using-ccode/accelerators.md`, `reference/accelerator-states.md`, `cookbook/accelerated-artifact.md` |
| OpenAPI | `using-ccode/openapi-workflows.md`, `cookbook/openapi-docs.md` |
| Database | `using-ccode/database-workflows.md`, `reference/database-inspection.md`, `cookbook/database-schema.md` |
| Pruning | `using-ccode/prune.md` |

Use local `ccode <command> --help` when CLI support is uncertain. Missing support
files require initialization with the appropriate version, not handwritten API
declarations. Do not edit accelerator records to hide unresolved state.
