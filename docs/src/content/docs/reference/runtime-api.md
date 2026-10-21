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
The reference helpers remain runtime exports from the default module.

Specification-shaped types now live in `@ccode/openapi/raw`, with the same
`V3_0`, `V3_1`, and `V3_2` namespaces and `Document<V>` selection alias.
Existing raw consumers must change their import path; the `parseOpenAPIFrom*`
functions continue to return raw documents.

## Public utility modules

General string transformations are exported from `@ccode/string`, Go-specific
naming transformations are exported from `@ccode/go`, and TypeScript
declaration identifiers are exported from `@ccode/typescript`. Route path
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

TypeScript.toTypeIdentifier("user id"); // "UserId"
TypeScript.toValueIdentifier("user id"); // "userId"
TypeScript.toValueIdentifier("class"); // "class_"

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
export function parseReference(input: string): ReferenceParts;
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
or JSON Pointers. Pass a reference object's `$ref` field to parse it.

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
