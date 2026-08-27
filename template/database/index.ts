import type * as MariaDBTypes from "./mariadb";
import type * as MySQLTypes from "./mysql";
import type * as PostgreSQLTypes from "./postgresql";
import type * as SQLiteTypes from "./sqlite";

export * as MariaDB from "./mariadb";
export * as MySQL from "./mysql";
export * as PostgreSQL from "./postgresql";
export * as SQLite from "./sqlite";

export type Engine = "postgresql" | "mysql" | "mariadb" | "sqlite";

export type Inspection =
  | PostgreSQLTypes.Inspection
  | MySQLTypes.Inspection
  | MariaDBTypes.Inspection
  | SQLiteTypes.Inspection;

export interface InspectionByEngine {
  postgresql: PostgreSQLTypes.Inspection;
  mysql: MySQLTypes.Inspection;
  mariadb: MariaDBTypes.Inspection;
  sqlite: SQLiteTypes.Inspection;
}

export interface InspectOptions<E extends Engine> {
  expectedEngine: E;
}
