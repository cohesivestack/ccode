---
title: Database Workflows
description: Inspect a database from a TypeScript process and prepare schema data for generation.
---

Use database inspection when you want generated files to match a database that
already exists. Your TypeScript process connects to PostgreSQL, MySQL, MariaDB,
or SQLite, reads its tables and columns, prepares the information your
generator needs, and passes it to a template. Inspection is read-only: it does
not change the database.

## A process that inspects SQLite

This example uses a dynamic URL so the engine-specific option and runtime guard
are visible. The database URL can come from any input your process already
controls; do not print it because it may contain credentials.

```ts
import type { Context } from "@ccode/context";
import * as Database from "@ccode/database";

export default function main(ctx: Context) {
  const connectionURL: string = "sqlite://./data/application.db";
  const inspection = ctx.inspectDatabase(connectionURL, {
    expectedEngine: "sqlite",
  });

  if (!Database.SQLite.isInspection(inspection)) {
    throw new Error("expected a SQLite inspection");
  }

  const database = inspection.databases[0];
  if (!database) {
    throw new Error("the SQLite connection has no databases");
  }

  const tables = database.tables.map((table) => ({
    name: table.name,
    columns: table.columns.map((column) => ({
      name: column.name,
      type: column.type.declaredType,
      nullable: column.nullable,
    })),
  }));

  ctx.generate("templates/schema-summary.tpl", "generated/schema.json", {
    database: database.name,
    tables,
  });
}
```

The template can remain focused on presentation:

```jinja
{
  "database": "{{ data.database }}",
  "tables": [
  {% for table in data.tables %}
    {
      "name": "{{ table.name }}",
      "columns": [
      {% for column in table.columns %}
        {"name": "{{ column.name }}", "type": "{{ column.type }}", "nullable": {{ column.nullable | lower }}}{% if not loop.last %},{% endif %}
      {% endfor %}
      ]
    }{% if not loop.last %},{% endif %}
  {% endfor %}
  ]
}
```

## Choosing the return type

When the URL is a string literal with a supported prefix, TypeScript infers the
engine-specific model:

```ts
const inspection = ctx.inspectDatabase("postgresql://localhost/application");
// Database.PostgreSQL.Inspection
```

When the URL is held in a general `string`, pass `expectedEngine` to get a
precise return type and to verify the configuration at runtime:

```ts
const connectionURL: string = "postgresql://user:password@localhost/application";
const inspection = ctx.inspectDatabase(connectionURL, {
  expectedEngine: "postgresql",
});
// Database.PostgreSQL.Inspection
```

If the URL scheme and `expectedEngine` disagree, inspection fails before the
database connection is opened. See [Database Inspection](/reference/database-inspection/)
for the returned model shapes and [Runtime API](/reference/runtime-api/) for
the complete method contract.

## Connection URLs

The supported prefixes are:

| Engine | Prefixes |
| --- | --- |
| PostgreSQL | `postgres://`, `postgresql://` |
| MySQL | `mysql://` |
| MariaDB | `maria://`, `mariadb://` |
| SQLite | `sqlite://` |

The URL is passed to the database inspector as provided. Make sure the process
has network access and credentials for server databases, and keep credentials
out of source control, generated files, and logs.
