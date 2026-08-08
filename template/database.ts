import type { MariaDBInspection } from "./database-mariadb";
import type { MySQLInspection } from "./database-mysql";
import type { PostgreSQLInspection } from "./database-postgresql";
import type { SQLiteInspection } from "./database-sqlite";

export * from "./database-mariadb";
export * from "./database-mysql";
export * from "./database-postgresql";
export * from "./database-sqlite";

export type DatabaseEngine =
  | "postgresql"
  | "mysql"
  | "mariadb"
  | "sqlite";

export type DatabaseInspection =
  | PostgreSQLInspection
  | MySQLInspection
  | MariaDBInspection
  | SQLiteInspection;

export interface DatabaseInspectionByEngine {
  postgresql: PostgreSQLInspection;
  mysql: MySQLInspection;
  mariadb: MariaDBInspection;
  sqlite: SQLiteInspection;
}

export type PostgreSQLConnectionURL =
  | `postgres://${string}`
  | `postgresql://${string}`;
export type MySQLConnectionURL = `mysql://${string}`;
export type MariaDBConnectionURL =
  | `maria://${string}`
  | `mariadb://${string}`;
export type SQLiteConnectionURL = `sqlite://${string}`;

export interface InspectDatabaseOptions<Engine extends DatabaseEngine> {
  expectedEngine: Engine;
}

export function isPostgreSQLInspection(
  inspection: DatabaseInspection,
): inspection is PostgreSQLInspection {
  return inspection.engine === "postgresql";
}

export function isMySQLInspection(
  inspection: DatabaseInspection,
): inspection is MySQLInspection {
  return inspection.engine === "mysql";
}

export function isMariaDBInspection(
  inspection: DatabaseInspection,
): inspection is MariaDBInspection {
  return inspection.engine === "mariadb";
}

export function isSQLiteInspection(
  inspection: DatabaseInspection,
): inspection is SQLiteInspection {
  return inspection.engine === "sqlite";
}
