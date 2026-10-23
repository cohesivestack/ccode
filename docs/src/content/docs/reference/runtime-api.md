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
import type * as OpenAPI from "@ccode/openapi";
import type * as OpenAPIRaw from "@ccode/openapi/raw";
import * as Database from "@ccode/database";

interface Context {
  println(message: string): void;
  runProcess(process: string): void;
  renderTemplate(templatePath: string, data: any): string;
  generate(templatePath: string, filePath: string, data: any): void;
  accelerate(
    id: string,
    templatePath: string,
    data: any,
    guidePath?: string,
  ): void;
  parseJSONFromBytes(jsonBytes: number[]): Record<string, any>;
  parseJSONFromString(jsonString: string): Record<string, any>;
  parseJSONFromFile(filePath: string): Record<string, any>;
  parseOpenAPIFromBytes(specBytes: number[]): OpenAPIRaw.Document;
  parseOpenAPIFromString(spec: string): OpenAPIRaw.Document;
  parseOpenAPIFromFile(filePath: string): OpenAPIRaw.Document;
  parseOpenAPIFromFile<V extends OpenAPIRaw.Version>(
    filePath: string,
    options: { expectedVersion: V },
  ): OpenAPIRaw.Document<V>;
  loadOpenAPIFromBytes(specBytes: number[]): OpenAPI.ApiModel;
  loadOpenAPIFromBytes<V extends OpenAPI.Version>(
    specBytes: number[], options: { expectedVersion: V },
  ): OpenAPI.ApiModel<V>;
  loadOpenAPIFromString(spec: string): OpenAPI.ApiModel;
  loadOpenAPIFromString<V extends OpenAPI.Version>(
    spec: string, options: { expectedVersion: V },
  ): OpenAPI.ApiModel<V>;
  loadOpenAPIFromFile(filePath: string): OpenAPI.ApiModel;
  loadOpenAPIFromFile<V extends OpenAPI.Version>(
    filePath: string, options: { expectedVersion: V },
  ): OpenAPI.ApiModel<V>;
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

The default `@ccode/openapi` entry point exposes the ergonomic `ApiModel`,
`ApiModelByVersion`, and `Version` types. Independent class declarations live in
`OpenAPI.V3_0`, `OpenAPI.V3_1`, and `OpenAPI.V3_2`. These are type contracts for
Go-created objects, not JavaScript constructors; use `import type` for models.

Specification-shaped types now live in `@ccode/openapi/raw`, with the same
`V3_0`, `V3_1`, and `V3_2` namespaces and `Document<V>` selection alias.
Existing raw consumers must change their import path; the `parseOpenAPIFrom*`
functions continue to return raw documents.

## Public utility modules

General string transformations are exported from `@ccode/string`, Go-specific
naming transformations and primitive type inference are exported from `@ccode/go`, and TypeScript
declaration identifiers and primitive type inference are exported from `@ccode/typescript`. Route path
transformations are exported from `@ccode/path`:

```ts
import * as Go from "@ccode/go";
import * as Path from "@ccode/path";
import * as Strings from "@ccode/string";
import * as TypeScript from "@ccode/typescript";

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
Go.toType("integer", { format: "int64" }); // "int64"

TypeScript.toTypeIdentifier("user id"); // "UserId"
TypeScript.toValueIdentifier("user id"); // "userId"
TypeScript.toValueIdentifier("class"); // "class_"
TypeScript.toType("integer", { format: "int64" }); // "bigint"

Path.toCurlyBrackets("/users/:userId"); // "/users/{userId}"
Path.toCurlyBrackets("/users/[userId]"); // "/users/{userId}"
Path.toColon("/users/{userId}"); // "/users/:userId"
Path.toSquareBrackets("/users/{userId}"); // "/users/[userId]"
```

Each path helper recognizes curly-bracket (`{id}`), colon (`:id`),
square-bracket (`[id]`), angle-bracket (`<id>`), and dollar (`$id`) parameters,
including mixed syntax in one path. It converts them to the requested format
and preserves the leading slash unless `omitLeadingSlash` is enabled:

```ts
Path.toAngleBrackets("/users/{userId}"); // "/users/<userId>"
Path.toDollar("/users/{userId}", {
  omitLeadingSlash: true,
}); // "users/$userId"
```

Parameter names start with a Unicode letter or underscore and may contain
letters, digits, combining marks, underscores, and hyphens. Dots separate
parameters, so `Path.toCurlyBrackets("/files/:file.:extension")` returns
`"/files/{file}.{extension}"`. Static text, empty placeholders, and malformed
bracket groups are preserved. Framework-specific regex constraints, optional
parameters, and catch-all semantics are outside the supported syntax.

These helpers only convert route path syntax. They do not interpolate runtime
values or validate parameter declarations. Run `ccode init` to refresh support
files after upgrading, and import path helpers from `@ccode/path`.

`camelCase`, `pascalCase`, `titleCase`, and `sentenceCase` accept an optional
read-only initialism array. The Go and TypeScript identifier helpers accept the
same option:

```ts
const initialisms = ["API", "ID"] as const;

Strings.pascalCase("api response id", initialisms); // "APIResponseID"
Go.toExportedIdentifier("api response id", initialisms); // "APIResponseID"
TypeScript.toTypeIdentifier("api response id", initialisms); // "APIResponseID"
TypeScript.toValueIdentifier("api response id", initialisms); // "apiResponseID"
```

Generic string and TypeScript identifier functions do not assume any
initialisms when the array is omitted. Go identifier functions always recognize
the conventional Go initialisms, including `API`, `HTTP`, `ID`, `JSON`, `SQL`,
`URL`, and `UUID`. Values supplied to a Go function extend that set and can
override the spelling of a default:

```ts
Strings.pascalCase("http server id"); // "HttpServerId"
Go.toExportedIdentifier("http server id"); // "HTTPServerID"
Go.toExportedIdentifier("user id", ["Id"]); // "UserId"
TypeScript.toTypeIdentifier("graph ql client", ["GraphQL"]); // "GraphQLClient"
```

Use TypeScript type identifiers for types, interfaces, classes, enums,
namespaces, and type aliases. Use value identifiers for variables, functions,
methods, parameters, local constants, and intentionally normalized properties.
TypeScript visibility is expressed with syntax such as `export`, `private`, and
`protected`, not identifier capitalization. Value identifiers that would be
reserved ECMAScript or strict-mode binding names receive a trailing underscore.
Use `@ccode/string` for generic casing and file-name formatting.

Initialism matching is case-insensitive, uses complete words or adjacent word
sequences, and preserves the spelling provided in the array. Each entry must be
a non-blank string; invalid inputs produce a runtime `TypeError`.

These transformations are pure from the TypeScript caller's perspective: they
return a new string and do not mutate their inputs or shared configuration.
They are imported modules, not members of `Context`.

## Go primitive type inference

`Go.toType` maps a primitive metadata type name to a Go type expression. The
Gonja `goType` filter uses the same implementation and options.

```ts
import * as Go from "@ccode/go";

export interface TypeOptions {
  readonly format?: string;
  readonly unsigned?: boolean;
  readonly source?: "openapi" | "postgresql" | "mysql" | "mariadb" | "sqlite";
  readonly overrides?: Readonly<Record<string, string>>;
  readonly fallback?: string | null;
}

// Exported as Go.TypeOptions and Go.toType:
// toType(value: string, options?: TypeOptions): string

Go.toType("boolean"); // "bool"
Go.toType("integer", { source: "openapi", format: "int64" }); // "int64"
Go.toType("bigint", { source: "mysql", unsigned: true }); // "uint64"
Go.toType("string", { format: "date-time" }); // "time.Time"
Go.toType("jsonb"); // "json.RawMessage"
```

Names, format annotations, source names, and override keys are matched
case-insensitively after trimming and collapsing whitespace. Output expressions
preserve their spelling. Numeric SQL size and precision modifiers are accepted,
including `varchar(255)`, `numeric(10, 2)`, and `timestamp(6) with time zone`.
A trailing SQL `UNSIGNED` selects the unsigned integer counterpart, except in
the SQLite profile. Size modifiers do not infer booleans, narrow integer widths,
or select a floating-point representation for exact decimals.

The default matches, without a source profile, are:

| Input names | Go expression |
| --- | --- |
| `bool`, `boolean` | `bool` |
| `string`, `text`, `varchar`, `nvarchar`, `char`, `nchar`, `character`, `character varying`, `tinytext`, `mediumtext`, `longtext`, `clob`, `uuid` | `string` |
| `int`, `integer` | `int` |
| `int8` | `int8` |
| `int16`, `short`, `smallint` | `int16` |
| `int32` | `int32` |
| `int64`, `long`, `bigint` | `int64` |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr` | Same name |
| `byte`, `rune`, `complex64`, `complex128`, `error` | Same name |
| `float32`, `float`, `real` | `float32` |
| `float64`, `double`, `double precision`, `number` | `float64` |
| `any`, `interface{}`, `interface {}` | `any` |
| `date`, `datetime`, `date-time`, `timestamp`, `timestamptz`, `time`, `time.Time`, `timetz`, `timestamp with time zone`, `timestamp without time zone`, `time with time zone`, `time without time zone` | `time.Time` |
| `binary`, `varbinary`, `blob`, `bytea`, `tinyblob`, `mediumblob`, `longblob` | `[]byte` |
| `json`, `jsonb` | `json.RawMessage` |

Source profiles take precedence over those defaults:

| Source | Input names | Go expression |
| --- | --- | --- |
| PostgreSQL | `smallint`, `int2`, `smallserial`, `serial2` | `int16` |
| PostgreSQL | `integer`, `int`, `int4`, `serial`, `serial4` | `int32` |
| PostgreSQL | `bigint`, `int8`, `bigserial`, `serial8` | `int64` |
| PostgreSQL | `real`, `float4` | `float32` |
| PostgreSQL | `float`, `float8`, `double precision` | `float64` |
| PostgreSQL | `bpchar` | `string` |
| MySQL / MariaDB | `tinyint` | `int8` |
| MySQL / MariaDB | `smallint` | `int16` |
| MySQL / MariaDB | `mediumint`, `int`, `integer` | `int32` |
| MySQL / MariaDB | `bigint` | `int64` |
| MySQL / MariaDB | `float` | `float32` |
| MySQL / MariaDB | `double`, `double precision`, `real` | `float64` |
| MySQL / MariaDB | `enum` | `string` |
| SQLite | `integer`, `int`, `tinyint`, `smallint`, `mediumint`, `bigint`, `int2`, `int8`, `unsigned big int` | `int64` |
| SQLite | `real`, `float`, `double`, `double precision` | `float64` |

`source: "openapi"` uses the generic mappings with format refinements. Source
profiles are inference conventions, not schema validators or database driver
configuration. In particular, SQLite declarations do not guarantee runtime value
types; MySQL/MariaDB `real` uses the double-precision convention, and
`tinyint(1)` remains `int8`. Use overrides for different project conventions.

Formats refine compatible base names:

| Base name | Format | Go expression |
| --- | --- | --- |
| `integer`, `number` | `int32` | `int32` |
| `integer`, `number` | `int64` | `int64` |
| `number` | `float` | `float32` |
| `number` | `double` | `float64` |
| `string` | `date`, `date-time`, `time` | `time.Time` |
| `string` | `byte`, `binary` | `[]byte` |

Unknown or incompatible formats retain the recognized base mapping. Formats
alone cannot make an unknown type recognizable. `unsigned: true` changes a
resolved `int`, `int8`, `int16`, `int32`, or `int64` to its unsigned counterpart;
other resolved types are unchanged.

Overrides are applied before source, format, and unsigned rules. Keys match the
normalized original input, then the base name after SQL modifiers are removed.
Conflicting override keys that normalize to the same name are rejected. Overrides
and fallback values must be non-blank strings; they are returned verbatim and
are not checked against a Go package or compiler.

```ts
const options: Go.TypeOptions = {
  source: "postgresql",
  overrides: { numeric: "decimal.Decimal" },
};

Go.toType("NUMERIC(10, 2)", options); // "decimal.Decimal"
Go.toType("long", { overrides: { long: "int" } }); // "int"
Go.toType("custom", { fallback: "any" }); // "any"
Go.toType("custom"); // throws TypeError
Go.toType("custom", { fallback: null }); // throws TypeError
```

Omitted or `undefined` TypeScript options use defaults. `fallback: null` is the
default strict behavior. Invalid inputs, option names, sources, and option value
types produce a catchable `TypeError`; a fallback does not hide configuration
errors or blank input names. Unsupported types such as exact `decimal`/`numeric`,
`money`, `interval`, spatial types, and project-defined names need an override or
an explicit fallback.

The helper returns type expressions only. The generator manages imports,
containers, named declarations, references, nullability, and field tags. Array, map, pointer, and union expressions are not parsed, and schemas are
not traversed. Date/time and
binary mappings do not install serialization or database codecs. The caller
selects any required conversion for date-only values, time-only values, or
engine-specific temporal encodings.

For an ergonomic OpenAPI primitive schema, use `schema.types[0]` after narrowing
`schema.kind` and verifying that it has one declared primitive type. Pass
`schema.format` as the format option. For MySQL/MariaDB columns, use
`column.type.name` and `column.type.unsigned`; for PostgreSQL use
`column.type.name`; for SQLite use `column.type.declaredType`. The source profile
is supplied explicitly by the caller.

## TypeScript primitive type inference

`TypeScript.toType` maps a primitive metadata name to a TypeScript type
expression. Its exported `TypeScript.TypeOptions` has the same fields as
`Go.TypeOptions`: `format`, `source`, `unsigned`, `overrides`, and `fallback`.

```ts
import * as TypeScript from "@ccode/typescript";

TypeScript.toType("boolean"); // "boolean"
TypeScript.toType("integer", { source: "openapi", format: "int64" }); // "bigint"
TypeScript.toType("string", { format: "date-time" }); // "Date"
TypeScript.toType("string", { format: "binary" }); // "Uint8Array"
TypeScript.toType("jsonb"); // "unknown"

const options: TypeScript.TypeOptions = {
  source: "postgresql",
  overrides: { numeric: "Decimal" },
};
TypeScript.toType("NUMERIC(10, 2)", options); // "Decimal"
TypeScript.toType("long", { overrides: { long: "number" } }); // "number"
TypeScript.toType("custom", { fallback: "unknown" }); // "unknown"
TypeScript.toType("custom"); // throws TypeError
TypeScript.toType("custom", { fallback: null }); // throws TypeError
```

The default mappings, without a source profile, are:

| Input names | TypeScript expression |
| --- | --- |
| `bool`, `boolean` | `boolean` |
| `string`, `text`, `varchar`, `nvarchar`, `char`, `nchar`, `character`, `character varying`, `tinytext`, `mediumtext`, `longtext`, `clob`, `uuid` | `string` |
| `int`, `integer`, `short`, `smallint`, `int8`, `int16`, `int32`, `uint`, `uint8`, `uint16`, `uint32`, `byte` | `number` |
| `long`, `bigint`, `int64`, `uint64` | `bigint` |
| `number`, `float`, `double`, `real`, `float32`, `float64`, `double precision` | `number` |
| `date`, `datetime`, `date-time`, `timestamp`, `timestamptz`, `time`, `timetz`, `timestamp with time zone`, `timestamp without time zone`, `time with time zone`, `time without time zone` | `Date` |
| `binary`, `varbinary`, `blob`, `bytea`, `tinyblob`, `mediumblob`, `longblob`, `Uint8Array` | `Uint8Array` |
| `json`, `jsonb` | `unknown` |
| `any`, `unknown`, `never`, `void`, `null`, `undefined`, `symbol` | Same name |

Source profiles take precedence over those defaults:

| Source | Input names | TypeScript expression |
| --- | --- | --- |
| PostgreSQL | `smallint`, `int2`, `smallserial`, `serial2`, `integer`, `int`, `int4`, `serial`, `serial4` | `number` |
| PostgreSQL | `bigint`, `int8`, `bigserial`, `serial8` | `bigint` |
| PostgreSQL | `real`, `float`, `float4`, `float8`, `double precision` | `number` |
| PostgreSQL | `bpchar` | `string` |
| MySQL / MariaDB | `tinyint`, `smallint`, `mediumint`, `int`, `integer`, `float`, `double`, `double precision`, `real` | `number` |
| MySQL / MariaDB | `bigint` | `bigint` |
| MySQL / MariaDB | `enum` | `string` |
| SQLite | `integer`, `int`, `tinyint`, `smallint`, `mediumint`, `bigint`, `int2`, `int8`, `unsigned big int` | `bigint` |
| SQLite | `real`, `float`, `double`, `double precision` | `number` |

SQLite integer declarations use `bigint` to cover their 64-bit range. MySQL
`tinyint(1)` stays `number`; boolean inference requires an explicit override.
`unsigned` is accepted but does not change TypeScript output: signed and unsigned
64-bit integers both map to `bigint`. `source: "openapi"` uses the generic
mappings with format refinements:

| Base name | Format | TypeScript expression |
| --- | --- | --- |
| `integer`, `number` | `int32` | `number` |
| `integer`, `number` | `int64` | `bigint` |
| `number` | `float`, `double` | `number` |
| `string` | `date`, `date-time`, `time` | `Date` |
| `string` | `byte`, `binary` | `Uint8Array` |

Matching follows the same normalization and precedence as `Go.toType`: trim and
collapse whitespace, match case-insensitively, strip numeric SQL parameters and
trailing `UNSIGNED`, and apply overrides before source and format rules. An
override first matches the normalized original name, then the SQL base name.
Unknown or incompatible formats retain the recognized base mapping.

Unknown types throw a catchable `TypeError` by default (`fallback: null`). Exact
decimal, money, interval, spatial, and custom types require an override or an
explicit fallback. Invalid inputs and options still throw when a fallback is
configured. Omitted or `undefined` options use defaults. Override and fallback
expressions must be non-blank strings and are returned verbatim.

The result is a type expression. The generator handles schema traversal,
references, containers, optional properties, nullability, imports, and
serialization. Literal `null` and `undefined` inputs map to those type names;
they do not infer nullability for another type. `Date`, `bigint`, and
`Uint8Array` mappings do not decode transport values or install codecs. Use
overrides when your runtime represents these values differently.

## println

Writes to process stdout.

```ts
ctx.println("building docs");
```

## renderTemplate

Renders a Gonja template and returns a string.

```ts
const text = ctx.renderTemplate("templates/readme.tpl", model);
```

All three output functions use the same Goja-to-Gonja converter. It preserves
object order, caches shared objects and cycles, and exposes synchronous model
methods and JavaScript functions as template callables. Methods retain their
original receiver, even when assigned to a template variable. No JSON
serialization is involved.

```ts
const api = ctx.loadOpenAPIFromFile("specs/api.yaml", { expectedVersion: "3.1" });
ctx.generate("templates/api.tpl", "generated/api.md", api);
```

```jinja
{% set operation = data.getOperation("get", "/users/{id}") %}
{% if operation %}
{{ operation.method | upper }} {{ operation.path }}
{% set parameter = operation.getParameter("path", "id") %}
{% if parameter %}Parameter: {{ parameter.name }}{% endif %}
{% endif %}

{% for server in data.servers %}
{{ server.expandUrl() }}
{% endfor %}
```

Calls accept positional arguments, including dictionaries and lists. A server
that declares a `region` variable can be expanded with
`server.expandUrl({"region": "eu"})`. Keyword arguments such as
`expandUrl(region="eu")` are rejected; pass one dictionary argument instead.
Missing or null call results become `none`, and exceptions become rendering
errors. Existing model objects passed into methods retain their original Goja
identity. New template dictionaries and lists are converted to Goja values.

Conversion is eager and scoped to each render. Each object is adapted once,
including objects reached by method results. Converted fields are snapshots;
mutating JavaScript data inside a template call does not refresh already
converted fields. Model methods are intended for read-only navigation.
Raw documents and ordinary plain data continue to use the same converter.

Normal field access, indexing, loops, and method calls can navigate cycles.
Whole-graph serialization and recursive filters are not cycle-aware; project
acyclic fields first when using those operations. Prefer projections for large
models when a template only needs a small portion of their data.

## generate

Renders a template and writes it to an output file.

```ts
ctx.generate("templates/readme.tpl", "README.generated.md", model);
```

## accelerate

Renders a template, tracks generated state, and writes safely to `output_path/<id>`.

```ts
ctx.accelerate("src/handlers.ts", "templates/handlers.tpl", model, "guides/handlers.md");
```

## parseJSONFrom*

Parses JSON and returns a JavaScript object. The root JSON value must be an object.

```ts
const settings = ctx.parseJSONFromFile("data/settings.json");
```

## loadOpenAPIFrom*

Loads an OpenAPI 3.0, 3.1, or 3.2 document on demand and constructs its ergonomic
model directly in Go. The synchronous call completes before returning. There is
no document cache across loader calls. File paths are relative to `ccode_path`.
An explicit `expectedVersion` selects the TypeScript return type and is checked
at runtime. String and byte loaders also accept this option.

```ts
import type * as OpenAPI from "@ccode/openapi";
import type * as OpenAPIRaw from "@ccode/openapi/raw";

const api: OpenAPI.V3_1.ApiModel = ctx.loadOpenAPIFromFile("specs/api.yaml", {
  expectedVersion: "3.1",
});

for (const server of api.servers) {
  for (const route of server.paths) {
    for (const operation of route.operations) {
      ctx.println(`${server.expandUrl()} ${operation.method} ${route.path}`);
    }
  }
}

const operation = api.getOperation("get", "/users/{id}");
const parameter = operation?.getParameter("path", "id");
const schema = operation?.getResponse("200")
  ?.getMediaType("application/json")?.schema;
const target = schema?.kind === "schema" ? schema.reference?.target : undefined;
const raw: OpenAPI.V3_1.DeepReadonly<OpenAPIRaw.V3_1.Document> = api.raw;
```

Collections preserve source order and are empty when undeclared, except where
absence carries meaning (`declaredServers`, `declaredSecurity`, schema `enum`).
Operations, route views, and server views share object identity. Returned graph
objects, collections, extensions, and original raw declarations are frozen.
The graph contains cycles. It can be passed directly to templates for normal
navigation and method calls. Project the required fields before JSON encoding
or using recursive serialization filters.

- `api.operations` contains ordinary routes; webhooks and callbacks have their
  own collections and ownership links. Recursive callbacks expose an explicit
  `recursiveTarget` edge and an empty `paths` array at the recursion boundary.
- `operation.parameters` merges path and operation parameters by exact location
  and name. Original declarations remain in `pathItem.parameters` and
  `operation.declaredParameters`.
- Operation servers override path servers, which override root servers. Scoped
  declarations remain distinct even when their URLs match. A synthetic `/`
  server represents the default. An explicit empty override uses this default.
- `server.paths` contains bindings filtered to the operations applicable to that
  server; `binding.pathItem` still points at the complete ordinary route.
- `expandUrl` substitutes defaults and validates variable overrides. Relative
  URLs stay relative; it does not append the operation path.
- Security alternatives inherit from the root unless explicitly overridden.
  An empty security array removes inherited requirements; an empty requirement
  object is an anonymous alternative. Schemes within an alternative use AND.
- `getPath` uses exact templates. `getResponse` and `getMediaType` use exact
  declared keys; response lookup does not apply range/default fallback.
- Schemas preserve composition. `getProperty` returns directly declared
  properties. Static references expose `reference.target`, including cycles;
  modern schema sibling constraints stay on the referring schema. Dynamic
  references retain their text because their targets depend on evaluation.
- OpenAPI 3.0 retains `nullable` and boolean exclusive bounds. Later versions
  retain type arrays, numeric exclusive bounds, boolean schemas, and dialect
  metadata. In 3.0, boolean constraints are available for `additionalProperties`.
- OpenAPI 3.2 exposes QUERY, additional methods, streaming item schemas, expanded
  encoding metadata, and the additional example and OAuth fields.

`ApiSource` identifies physical source URIs, JSON Pointers, and one-based source
positions. `ApiReference` records the original reference, its resolved URI, and
both source and target locations. `raw` is the original parsed declaration,
without materialized references; it does not retain YAML comments or formatting.
This differs from the materialized raw result returned by the existing file
parser.

Local file references, JSON Pointers, and indexed schema resources/anchors are
supported. In-memory documents support internal references; use the file loader
for relative external files. Unindexed remote resources are not fetched. The
loader uses libopenapi for structural model parsing and checks navigation
ambiguities such as duplicate operation IDs and parameters; it is not a full
OpenAPI or JSON Schema validator. Custom vocabulary keywords remain in `raw`.

Run `ccode init` to refresh an existing workspace's support files after upgrading.

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
