---
title: Database Schema
description: Generate a simple schema report from a SQLite database.
---

This cookbook shows a complete process that inspects a SQLite database and
generates a Markdown table report. The same workflow applies to PostgreSQL,
MySQL, and MariaDB; only the returned model shape changes.

## Layout

```text
ccode/
  database/
    schema-report.ts
  templates/
    schema-report.tpl
```

## Process

```ts
import type { Context } from "@ccode/context";
import * as Database from "@ccode/database";

export default function main(ctx: Context) {
  const inspection = ctx.inspectDatabase("sqlite://./data/application.db");
  if (!Database.SQLite.isInspection(inspection)) {
    throw new Error("expected a SQLite inspection");
  }

  const database = inspection.databases[0];
  if (!database) {
    throw new Error("database has no schema");
  }

  ctx.generate("templates/schema-report.tpl", "generated/schema.md", {
    database,
  });
}
```

## Template

```jinja
# {{ data.database.name }}

{% for table in data.database.tables %}
## {{ table.name }}

| Column | Type | Nullable |
| --- | --- | --- |
{% for column in table.columns %}
| `{{ column.name }}` | `{{ column.type.declaredType }}` | {{ column.nullable }} |
{% endfor %}
{% endfor %}
```

## Run

```bash
ccode run database/schema-report
```

The process owns schema traversal and the template only formats the prepared
data. For PostgreSQL, use `inspection.database.schemas`; for MySQL, MariaDB,
and SQLite, use `inspection.databases`. See [Database Inspection](/reference/database-inspection/)
for the engine-specific models.
