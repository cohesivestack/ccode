---
title: Runtime API
description: The TypeScript modules and Context API exposed to Cohesive Code processes.
---

Import `Context` from the generated workspace support files:

```ts
import type { Context } from "@ccode/context";
```

## Surface

```ts
import * as OpenAPI from "@ccode/openapi";
import * as Database from "@ccode/database";

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
  parseOpenAPIFromBytes(specBytes: number[]): OpenAPI.Document;
  parseOpenAPIFromString(spec: string): OpenAPI.Document;
  parseOpenAPIFromFile(filePath: string): OpenAPI.Document;
  parseOpenAPIFromFile<V extends OpenAPI.Version>(
    filePath: string,
    options: { expectedVersion: V },
  ): OpenAPI.Document<V>;
  inspectDatabase(
    connectionURL: Database.PostgreSQL.ConnectionURL,
  ): Database.PostgreSQL.Inspection;
  inspectDatabase(
    connectionURL: Database.MySQL.ConnectionURL,
  ): Database.MySQL.Inspection;
  inspectDatabase(
    connectionURL: Database.MariaDB.ConnectionURL,
  ): Database.MariaDB.Inspection;
  inspectDatabase(
    connectionURL: Database.SQLite.ConnectionURL,
  ): Database.SQLite.Inspection;
  inspectDatabase<Engine extends Database.Engine>(
    connectionURL: string,
    options: Database.InspectOptions<Engine>,
  ): Database.InspectionByEngine[Engine];
  inspectDatabase(connectionURL: string): Database.Inspection;
}
```

Use the local generated `context.ts` as the final contract for the installed workspace version.

General OpenAPI types are top-level exports from `@ccode/openapi`, including
`Version`, `Document`, `DocumentByVersion`, `Operation`, `Parameter`,
`Parameters`, `Request`, `ReferenceLike`, and `ReferenceParts`. Normative
version-specific declarations are ES modules exposed as `V3_0`, `V3_1`,
and `V3_2`. The module also exports `isReference` and `parseReference` for
working with preserved `$ref` values.

## Public utility modules

General string transformations are exported from `@ccode/string`, while
Go-specific naming transformations are exported from `@ccode/go`. OpenAPI path
transformations are available under `OpenAPI.Path`:

```ts
import * as Go from "@ccode/go";
import * as Strings from "@ccode/string";

Strings.camelCase("user name"); // "userName"
Strings.pascalCase("user name"); // "UserName"
Strings.snakeCase("User Name"); // "user_name"
Strings.kebabCase("User Name"); // "user-name"
Strings.constantCase("User Name"); // "USER_NAME"
Strings.dotCase("User Name"); // "user.name"
Strings.pathCase("User Name"); // "user/name"
Strings.titleCase("user name"); // "User Name"
Strings.sentenceCase("user name"); // "User name"
Strings.upperFirst("user name"); // "User name"
Strings.lowerFirst("User Name"); // "user Name"
Strings.normalizeSpace(" user   name "); // "user name"

Go.toExportedIdentifier("user_id"); // "UserID"
Go.toUnexportedIdentifier("user_id"); // "userID"
Go.toPackageName("HTTP Utils"); // "httputils"

OpenAPI.Path.toColon("/users/{userId}"); // "/users/:userId"
OpenAPI.Path.toSquareBrackets("/users/{userId}"); // "/users/[userId]"
```

The path helpers convert every nonempty OpenAPI `{parameter}` expression to
colon, square-bracket, angle-bracket, or dollar syntax. They preserve the
leading slash unless `omitLeadingSlash` is enabled:

```ts
OpenAPI.Path.toAngleBrackets("/users/{userId}"); // "/users/<userId>"
OpenAPI.Path.toDollar("/users/{userId}", {
  omitLeadingSlash: true,
}); // "users/$userId"
```

The same functions can be imported directly from `@ccode/openapi/path`. These
helpers only convert path syntax. They do not interpolate runtime parameter
values or validate parameter declarations against a Path Item or Operation.

`camelCase`, `pascalCase`, `titleCase`, and `sentenceCase` accept an optional
read-only initialism array. `toExportedIdentifier` and
`toUnexportedIdentifier` accept the same option:

```ts
const initialisms = ["API", "ID"] as const;

Strings.pascalCase("api response id", initialisms); // "APIResponseID"
Go.toExportedIdentifier("api response id", initialisms); // "APIResponseID"
```

Generic string functions do not assume any initialisms when the array is
omitted. Go identifier functions always recognize the conventional Go
initialisms, including `API`, `HTTP`, `ID`, `JSON`, `SQL`, `URL`, and `UUID`.
Values supplied to a Go function extend that set and can override the spelling
of a default:

```ts
Strings.pascalCase("http server id"); // "HttpServerId"
Go.toExportedIdentifier("http server id"); // "HTTPServerID"
Go.toExportedIdentifier("user id", ["Id"]); // "UserId"
```

Initialism matching is case-insensitive, uses complete words or adjacent word
sequences, and preserves the spelling provided in the array. Each entry must be
a non-blank string; invalid inputs produce a runtime `TypeError`.

These transformations are pure from the TypeScript caller's perspective: they
return a new string and do not mutate their inputs or shared configuration.
They are imported modules, not members of `Context`.

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

## OpenAPI reference helpers

`parseReference` extracts portable parts from a local OpenAPI reference
without resolving or loading the referenced document:

```ts
export interface ReferenceLike {
  readonly $ref: string;
}

export interface ReferenceParts {
  readonly raw: string;
  readonly document: string;
  readonly directory: string;
  readonly directorySegments: readonly string[];
  readonly filename: string;
  readonly documentName: string;
  readonly fragment: string;
}

export function isReference(value: unknown): value is ReferenceLike;
export function parseReference(
  input: string | ReferenceLike,
): ReferenceParts;
```

```ts
import * as OpenAPI from "@ccode/openapi";

const source = OpenAPI.parseReference(
  "./paths/app/common/countries.yaml#/countries",
);

source.documentName; // "countries"
source.directorySegments; // ["paths", "app", "common"]
```

Calculated path and fragment fields are percent-decoded, while `raw` and
`document` preserve the original text. The helper accepts only the same local
reference subset as file parsing: no URI scheme or host, protocol-relative
remote reference, query string, malformed percent-encoding, or non-pointer
fragment. It validates and describes the reference but does not resolve files
or JSON Pointers.

## inspectDatabase

Inspects a PostgreSQL, MySQL, MariaDB, or SQLite database and returns the
corresponding Cohesive Code model. Literal connection URLs infer the engine from
their prefix:

```ts
const database = ctx.inspectDatabase("postgresql://localhost/app");
// Database.PostgreSQL.Inspection
```

For a connection URL held in a general `string`, provide `expectedEngine` when
an engine-specific return type is needed. The runtime verifies that it agrees
with the URL before inspecting the database.

```ts
const connectionURL: string = "mysql://user:password@localhost/app";
const database = ctx.inspectDatabase(connectionURL, {
  expectedEngine: "mysql",
});
// Database.MySQL.Inspection
```

The supported URL prefixes are `postgres://`, `postgresql://`, `mysql://`,
`maria://`, `mariadb://`, and `sqlite://`. Engine-specific models and narrowing
helpers are exported from `@ccode/database`:

```ts
import * as Database from "@ccode/database";

const database = ctx.inspectDatabase(connectionURL);
if (Database.SQLite.isInspection(database)) {
  ctx.println(database.databases[0]?.name ?? "no database");
}
```

The inspection call opens the database, reads the schema visible to that
connection, and closes the connection before returning. Connection failures and
schema inspection failures are runtime errors, so a process stops unless it
handles them explicitly. Keep credentials out of generated output and process
logs.
