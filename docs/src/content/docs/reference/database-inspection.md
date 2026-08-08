---
title: Database Inspection
description: TypeScript models returned by Cohesive Code database inspection.
---

`ctx.inspectDatabase` returns a Cohesive Code model tailored to the database
engine. The models are intentionally engine-specific so generators can use
native details without depending on an inspection library.

Import the shared types and narrowing helpers from the generated module:

```ts
import type { Context } from "@ccode/context";
import type {
  DatabaseInspection,
  PostgreSQLInspection,
  SQLiteInspection,
} from "@ccode/database";
import { isPostgreSQLInspection, isSQLiteInspection } from "@ccode/database";
```

## Top-level shapes

| Engine | Top-level schema collection | Relationships |
| --- | --- | --- |
| PostgreSQL | `inspection.database.schemas` | `inspection.database.relationships` |
| MySQL | `inspection.databases` | `inspection.relationships` |
| MariaDB | `inspection.databases` | `inspection.relationships` |
| SQLite | `inspection.databases` | `inspection.databases[n].relationships` |

Every inspection also has an `engine` discriminator. Use it directly or use a
type guard when the value is a `DatabaseInspection` union:

```ts
function listTables(inspection: DatabaseInspection): string[] {
  if (isPostgreSQLInspection(inspection)) {
    return inspection.database.schemas.flatMap((schema) =>
      schema.tables.map((table) => `${schema.name}.${table.name}`),
    );
  }

  return inspection.databases.flatMap((database) =>
    database.tables.map((table) => `${database.name}.${table.name}`),
  );
}
```

Properties marked optional in the generated TypeScript models are omitted from
the returned JSON when the database does not provide a value. Arrays are
returned as empty arrays when there are no items.

## Tables and columns

All four engines expose tables with `name`, `columns`, `primaryKey`, `indexes`,
and `foreignKeys`. A table's `primaryKey` is optional because a table does not
need to have one. Each column includes:

- `name` and one-based `position`.
- `nullable`.
- An engine-specific `type` object.
- Optional `defaultExpression` and `generatedExpression` when available.

The engine-specific table models also expose native metadata:

- PostgreSQL tables may have `comment`; PostgreSQL columns may have `comment`,
  `identity`, and type `schema`/`arrayDimensions` details.
- MySQL and MariaDB tables expose storage engine, character set, collation, and
  comments. Columns expose auto-increment, character set, collation, and
  comments.
- SQLite tables expose `strict` and `withoutRowID`.

## PostgreSQL

PostgreSQL returns one `database` object containing schemas. Each schema has
`name`, `tables`, and named `enumTypes`:

```ts
function enumValues(inspection: PostgreSQLInspection): string[] {
  return inspection.database.schemas.flatMap((schema) =>
    schema.enumTypes.flatMap((enumType) => enumType.values),
  );
}
```

PostgreSQL column types include `name`, `nativeType`, and optional
`schema`, `length`, `precision`, `scale`, and `arrayDimensions`.

Foreign-key and relationship table references include both `schema` and
`table`, which lets generators distinguish tables with the same name in
different schemas.

## MySQL and MariaDB

MySQL and MariaDB return a `databases` array because a server connection can
expose more than one database. Each database has `name`, optional character set
and collation, and `tables`.

Their column type models include native details such as `unsigned`,
`length`, `precision`, and `scale`. Inline custom values are represented on the
type itself:

- `enumValues` for an inline `ENUM` column.
- `setValues` for an inline `SET` column.

These engines do not expose PostgreSQL-style named enum type objects. Table
references identify the database and table.

MySQL indexes include `kind` and `visible`. MariaDB indexes include `kind` and
`ignored`; use the engine-specific model when these distinctions matter.

## SQLite

SQLite returns a `databases` array. Its type model preserves both the declared
type and SQLite's inferred affinity:

```ts
function printSQLiteTypes(ctx: Context, inspection: DatabaseInspection) {
  if (!isSQLiteInspection(inspection)) return;

  for (const database of inspection.databases) {
    for (const table of database.tables) {
      for (const column of table.columns) {
        ctx.println(`${column.name}: ${column.type.declaredType} (${column.type.affinity})`);
      }
    }
  }
}
```

The possible affinities are `integer`, `real`, `text`, `blob`, and `numeric`.
SQLite relationships belong to each database object rather than the inspection
root.

## Keys, indexes, and relationships

Primary keys expose their participating column names. Indexes expose whether
they are unique and their ordered parts; a part may refer to a column or an
expression and may be descending. Engine-specific index metadata is documented
by the generated types.

Foreign keys expose local columns, referenced table information, referenced
columns, and update/delete actions. PostgreSQL and SQLite additionally expose
deferrability fields.

Relationships summarize foreign-key connections with `fromTable`,
`fromColumns`, `toTable`, `toColumns`, `cardinality`, and `optional`. The
cardinality is currently `many-to-one` or `one-to-one`. The exact table
reference shape follows the engine's namespace rules described above.

For complete property names and literal unions, open the generated files under
`.ccode/lib/` or import the corresponding types from `@ccode/database`.
