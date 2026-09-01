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

## String, Go, TypeScript naming, and OpenAPI path filters

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
| `goPackage` | `"HTTP Utils"` → `"httputils"` |
| `typeScriptType` | `"user id"` → `"UserId"` |
| `typeScriptValue` | `"user id"` → `"userId"` |
| `openAPIPathToColon` | `"/users/{userId}"` → `"/users/:userId"` |
| `openAPIPathToSquareBrackets` | `"/users/{userId}"` → `"/users/[userId]"` |
| `openAPIPathToAngleBrackets` | `"/users/{userId}"` → `"/users/<userId>"` |
| `openAPIPathToDollar` | `"/users/{userId}"` → `"/users/$userId"` |

Use them with the normal Gonja pipe syntax:

```jinja
type {{ data.name | goExported }} struct {}

export interface {{ data.name | typeScriptType }} {
  {{ data.property | typeScriptValue }}: string;
}

{{ data.path | openAPIPathToColon }}
{{ data.path | openAPIPathToColon(omitLeadingSlash=true) }}
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
non-blank strings. The four OpenAPI path filters accept the keyword-only boolean
`omitLeadingSlash`, which defaults to `false`. Other filters do not accept
arguments.

OpenAPI path filters replace every well-formed, nonempty `{parameter}`
expression and leave malformed expressions unchanged. They only convert the
path syntax: they do not resolve parameter values or validate declarations
against a Path Item or Operation.

## Path rules

Template paths resolve relative to `ccode_path`.

```ts
ctx.renderTemplate("templates/readme.tpl", model);
ctx.generate("templates/readme.tpl", "README.generated.md", model);
ctx.accelerate("src/handlers.ts", "templates/handlers.tpl", model);
```

## Data rules

Pass plain objects, arrays, strings, numbers, booleans, and null-like values. Avoid class instances or values that depend on JavaScript prototypes because the runner converts values before giving them to the template engine.
