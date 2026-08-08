export interface MariaDBInspection {
  engine: "mariadb";
  databases: MariaDBDatabase[];
  relationships: MariaDBRelationship[];
}

export interface MariaDBDatabase {
  name: string;
  characterSet?: string;
  collation?: string;
  tables: MariaDBTable[];
}

export interface MariaDBTable {
  name: string;
  storageEngine: string;
  characterSet?: string;
  collation?: string;
  comment?: string;
  columns: MariaDBColumn[];
  primaryKey?: MariaDBPrimaryKey;
  indexes: MariaDBIndex[];
  foreignKeys: MariaDBForeignKey[];
}

export interface MariaDBColumn {
  name: string;
  position: number;
  type: MariaDBType;
  nullable: boolean;
  defaultExpression?: string;
  generatedExpression?: string;
  autoIncrement: boolean;
  characterSet?: string;
  collation?: string;
  comment?: string;
}

export interface MariaDBType {
  name: string;
  nativeType: string;
  length?: number;
  precision?: number;
  scale?: number;
  unsigned: boolean;
  enumValues?: string[];
  setValues?: string[];
}

export interface MariaDBPrimaryKey {
  name?: string;
  columns: string[];
}

export interface MariaDBIndex {
  name: string;
  unique: boolean;
  kind: "btree" | "hash" | "fulltext" | "spatial" | "other";
  parts: MariaDBIndexPart[];
  ignored: boolean;
}

export interface MariaDBIndexPart {
  column?: string;
  expression?: string;
  prefixLength?: number;
  descending: boolean;
}

export interface MariaDBForeignKey {
  name?: string;
  columns: MariaDBForeignKeyColumn[];
  referencedTable: MariaDBTableReference;
  onUpdate: MariaDBReferentialAction;
  onDelete: MariaDBReferentialAction;
}

export interface MariaDBForeignKeyColumn {
  column: string;
  referencedColumn: string;
}

export interface MariaDBTableReference {
  database: string;
  table: string;
}

export type MariaDBReferentialAction =
  | "no-action"
  | "restrict"
  | "cascade"
  | "set-null";

export interface MariaDBRelationship {
  fromTable: MariaDBTableReference;
  fromColumns: string[];
  toTable: MariaDBTableReference;
  toColumns: string[];
  foreignKey?: string;
  cardinality: "many-to-one" | "one-to-one";
  optional: boolean;
}
