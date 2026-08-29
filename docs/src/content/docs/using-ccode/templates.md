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

## String and Go naming filters

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

Use them with the normal Gonja pipe syntax:

```jinja
type {{ data.name | goExported }} struct {}
```

`camelCase`, `pascalCase`, `titleCase`, `sentenceCase`, `goExported`, and
`goUnexported` accept a keyword-only `initialisms` list:

```jinja
{{ data.name | pascalCase(initialisms=["API", "ID"]) }}
{{ data.name | goUnexported(initialisms=data.initialisms) }}
```

Generic case filters have no built-in initialisms. The Go filters include the
conventional Go initialism set, and a supplied list extends or overrides that
set. Initialisms match case-insensitively and preserve the supplied spelling.
The list must contain only non-blank strings. Other filters do not accept
arguments.

## Path rules

Template paths resolve relative to `ccode_path`.

```ts
ctx.renderTemplate("templates/readme.tpl", model);
ctx.generate("templates/readme.tpl", "README.generated.md", model);
ctx.accelerate("src/handlers.ts", "templates/handlers.tpl", model);
```

## Data rules

Pass plain objects, arrays, strings, numbers, booleans, and null-like values. Avoid class instances or values that depend on JavaScript prototypes because the runner converts values before giving them to the template engine.
