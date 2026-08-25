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
import type {
  OpenAPIDocument,
  OpenAPIVersion,
} from "@ccode/openapi";
import type {
  MariaDBConnectionURL,
  MariaDBInspection,
  MySQLConnectionURL,
  MySQLInspection,
  PostgreSQLConnectionURL,
  PostgreSQLInspection,
  SQLiteConnectionURL,
  SQLiteInspection,
  DatabaseInspection,
  DatabaseInspectionByEngine,
  DatabaseEngine,
  InspectDatabaseOptions,
} from "@ccode/database";

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
  parseOpenAPIFromFile<V extends OpenAPIVersion>(
    filePath: string,
    options: { expectedVersion: V },
  ): OpenAPIDocument<V>;
  inspectDatabase(
    connectionURL: PostgreSQLConnectionURL,
  ): PostgreSQLInspection;
  inspectDatabase(connectionURL: MySQLConnectionURL): MySQLInspection;
  inspectDatabase(connectionURL: MariaDBConnectionURL): MariaDBInspection;
  inspectDatabase(connectionURL: SQLiteConnectionURL): SQLiteInspection;
  inspectDatabase<Engine extends DatabaseEngine>(
    connectionURL: string,
    options: InspectDatabaseOptions<Engine>,
  ): DatabaseInspectionByEngine[Engine];
  inspectDatabase(connectionURL: string): DatabaseInspection;
}
```

Use the local generated `context.ts` as the final contract for the installed workspace version.

General OpenAPI types are top-level exports from `@ccode/openapi`, including
`OpenAPIVersion`, `OpenAPIDocument`, `OpenAPIDocumentByVersion`,
`OpenAPIOperation`, `OpenAPIParameter`, `OpenAPIParameters`, and
`OpenAPIRequest`. Version-specific declarations remain grouped under the
`OpenAPIV3`, `OpenAPIV3_1`, and `OpenAPIV3_2` namespaces.

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

Parses OpenAPI v3 input and returns a JSON-like object. `parseOpenAPIFromFile` resolves paths relative to `ccode_path` and recursively resolves internal and external file references by default. External paths are relative to the document containing the reference. Resolved objects retain their original `$ref`, so a referenced Path Item exposes both provenance such as `pathItem.$ref` and materialized operations such as `pathItem.get`.

```ts
const spec = ctx.parseOpenAPIFromFile("specs/api.yaml");

const spec31 = ctx.parseOpenAPIFromFile("specs/api.yaml", {
  expectedVersion: "3.1",
});

const countries = spec31.paths?.["/countries"];
if (countries?.$ref && countries.get) {
  console.log(countries.$ref, countries.get.operationId);
}
```

Without `expectedVersion`, the return type is the union of supported OpenAPI document versions. With `expectedVersion`, the runtime verifies the document's `openapi` field and TypeScript returns the corresponding version-specific document type.

## inspectDatabase

Inspects a PostgreSQL, MySQL, MariaDB, or SQLite database and returns the
corresponding Cohesive Code model. Literal connection URLs infer the engine from
their prefix:

```ts
const database = ctx.inspectDatabase("postgresql://localhost/app");
// PostgreSQLInspection
```

For a connection URL held in a general `string`, provide `expectedEngine` when
an engine-specific return type is needed. The runtime verifies that it agrees
with the URL before inspecting the database.

```ts
const connectionURL: string = "mysql://user:password@localhost/app";
const database = ctx.inspectDatabase(connectionURL, {
  expectedEngine: "mysql",
});
// MySQLInspection
```

The supported URL prefixes are `postgres://`, `postgresql://`, `mysql://`,
`maria://`, `mariadb://`, and `sqlite://`. Engine-specific models and narrowing
helpers are exported from `@ccode/database`:

```ts
import { isSQLiteInspection } from "@ccode/database";

const database = ctx.inspectDatabase(connectionURL);
if (isSQLiteInspection(database)) {
  ctx.println(database.databases[0]?.name ?? "no database");
}
```

The inspection call opens the database, reads the schema visible to that
connection, and closes the connection before returning. Connection failures and
schema inspection failures are runtime errors, so a process stops unless it
handles them explicitly. Keep credentials out of generated output and process
logs.
