---
title: Templates
description: Rendering with Gonja templates and template-friendly data.
---

Cohesive Code uses Gonja templates for text rendering. A template receives a single root variable named `data`.

## Basic template

Process:

```ts
ctx.generate("templates/greeting.tpl", "generated/greeting.txt", {
  name: "Cohesive Code",
});
```

Template:

```jinja
Hello {{ data.name }}!
```

Output:

```text
Hello Cohesive Code!
```

## Model navigation and method calls

Raw documents, ergonomic models, and ordinary JavaScript objects share one
converter. Pass an ergonomic model directly when the template needs its
navigation methods:

```ts
const api = ctx.loadOpenAPIFromFile("specs/api.yaml", { expectedVersion: "3.1" });
ctx.generate("templates/operations.tpl", "generated/operations.md", api);
```

```jinja
{% for server in data.servers %}
## {{ server.expandUrl() }}
{% for route in server.paths %}
{% for operation in route.operations %}
- {{ operation.method | upper }} {{ route.path }}
{% endfor %}
{% endfor %}
{% endfor %}

{% set operation = data.getOperationById("getUser") %}
{% if operation %}
{% set response = operation.getResponse("200") %}
{% if response %}{{ response.description }}{% endif %}
{% endif %}
```

Method arguments are positional. Options use dictionary syntax, for example
`server.expandUrl({"region": "eu"})` when that server declares `region`.
JavaScript exceptions become rendering errors. Conversion caches shared objects
and circular relationships for each render; it does not lazily load fields.
Avoid serializing or recursively filtering an entire cyclic graph. Use a plain
projection for those operations, or when only a few fields are needed.

## Keep templates shallow

Templates should format data. They should not own deep transformation logic, fallback naming, schema walking, or cross-file decisions.

Prefer shaping a stable model in TypeScript:

```ts
const operations = Object.entries(spec.paths ?? {}).flatMap(([path, item]) => {
  if (!item) return [];
  return ["get", "post", "put", "patch", "delete"].flatMap((method) => {
    const operation = item[method];
    if (!operation) return [];
    return [{
      method: method.toUpperCase(),
      path,
      operationId: operation.operationId ?? `${method}_${path.replace(/[^a-z0-9]+/gi, "_")}`,
    }];
  });
});
```

Then render with simple loops:

```jinja
{% for operation in data.operations %}
- `{{ operation.operationId }}` -> `{{ operation.method }} {{ operation.path }}`
{% endfor %}
```

## String, Go, TypeScript naming, and route path filters

Cohesive Code adds naming filters to Gonja's standard filters. All of these
filters require a string input:

| Filter | Example result |
| --- | --- |
| `camelCase` | `"HTTP server"` → `"httpServer"` |
| `pascalCase` | `"HTTP server"` → `"HttpServer"` |
| `snakeCase` | `"HTTP server"` → `"http_server"` |
| `kebabCase` | `"HTTP server"` → `"http-server"` |
| `constantCase` | `"HTTP server"` → `"HTTP_SERVER"` |
| `dotCase` | `"HTTP server"` → `"http.server"` |
| `pathCase` | `"HTTP server"` → `"http/server"` |
| `titleCase` | `"HTTP server"` → `"Http Server"` |
| `sentenceCase` | `"HTTP server"` → `"Http server"` |
| `upperFirst` | `"userAccount"` → `"UserAccount"` |
| `lowerFirst` | `"UserAccount"` → `"userAccount"` |
| `normalizeSpace` | `"  user   account  "` → `"user account"` |
| `goExported` | `"user id"` → `"UserID"` |
| `goUnexported` | `"user id"` → `"userID"` |
| `goType` | `"integer"` with `format="int64"` → `"int64"` |
| `goPackage` | `"HTTP Utils"` → `"httputils"` |
| `typeScriptType` | `"user id"` → `"UserId"` |
| `typeScriptValue` | `"user id"` → `"userId"` |
| `pathToCurlyBrackets` | `"/users/:userId"` → `"/users/{userId}"` |
| `pathToColon` | `"/users/{userId}"` → `"/users/:userId"` |
| `pathToSquareBrackets` | `"/users/{userId}"` → `"/users/[userId]"` |
| `pathToAngleBrackets` | `"/users/{userId}"` → `"/users/<userId>"` |
| `pathToDollar` | `"/users/{userId}"` → `"/users/$userId"` |

Use them with the normal Gonja pipe syntax:

```jinja
type {{ data.name | goExported }} struct {}

export interface {{ data.name | typeScriptType }} {
  {{ data.property | typeScriptValue }}: string;
}

{{ data.path | pathToColon }}
{{ data.path | pathToColon(omitLeadingSlash=true) }}
```

`camelCase`, `pascalCase`, `titleCase`, `sentenceCase`, `goExported`,
`goUnexported`, `typeScriptType`, and `typeScriptValue` accept a keyword-only
`initialisms` list:

```jinja
{{ data.name | pascalCase(initialisms=["API", "ID"]) }}
{{ data.name | goUnexported(initialisms=data.initialisms) }}
{{ data.name | typeScriptType(initialisms=["API", "ID"]) }}
{{ data.name | typeScriptValue(initialisms=data.initialisms) }}
```

Generic case filters have no built-in initialisms. The Go filters include the
conventional Go initialism set, and a supplied list extends or overrides that
set. TypeScript filters also have no built-in initialisms, so pass custom values
to preserve spellings such as `API`, `ID`, or `GraphQL`. TypeScript type names
use PascalCase and value names use camelCase; visibility remains controlled by
TypeScript syntax rather than capitalization. Reserved value binding names
receive a trailing underscore. Initialisms match case-insensitively and preserve
safe characters from the supplied spelling. The list must contain only
non-blank strings. The five route path filters accept the keyword-only boolean
`omitLeadingSlash`, which defaults to `false`. The `goType` filter accepts the
options described below. Filters outside these groups do not accept arguments.

Route path filters accept curly-bracket, colon, square-bracket, angle-bracket,
and dollar parameters, including mixed syntax, and preserve malformed groups.
They use the same parameter name rules as `@ccode/path`; see
[Runtime API](/reference/runtime-api/). They only convert path syntax and do not
resolve parameter values or validate declarations.

The `goType` filter maps primitive metadata names to Go expressions. It accepts
keyword-only `format`, `source`, `unsigned`, `overrides`, and `fallback`, using
the same defaults and matching rules as `Go.toType`:

```jinja
{{ "boolean" | goType }}
{{ "integer" | goType(source="openapi", format="int64") }}
{{ column.type.name | goType(source="mysql", unsigned=column.type.unsigned) }}
{{ "numeric" | goType(source="postgresql", overrides=data.typeOverrides) }}
{{ "unrecognized" | goType(fallback="any") }}
```

`overrides` can be a dictionary literal or a mapping supplied in template data.
Its keys and values must be non-blank strings. `fallback=none` is the default:
an unsupported type fails rendering with the filter name in the error. An
explicit non-blank fallback permits unknown names. Invalid options still fail
rendering. The filter does not infer containers or field optionality and does
not manage imports. See [Go primitive type inference](/reference/runtime-api/#go-primitive-type-inference)
for the full default and source-specific mappings.

## Path rules

Template paths resolve relative to `ccode_path`.

```ts
ctx.renderTemplate("templates/readme.tpl", model);
ctx.generate("templates/readme.tpl", "README.generated.md", model);
ctx.accelerate("src/handlers.ts", "templates/handlers.tpl", model);
```

## Data rules

Pass plain objects, arrays, strings, numbers, booleans, and null-like values. Avoid class instances or values that depend on JavaScript prototypes because the runner converts values before giving them to the template engine.
