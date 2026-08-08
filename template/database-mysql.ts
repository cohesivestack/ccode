export interface MySQLInspection {
  engine: "mysql";
  databases: MySQLDatabase[];
  relationships: MySQLRelationship[];
}

export interface MySQLDatabase {
  name: string;
  characterSet?: string;
  collation?: string;
  tables: MySQLTable[];
}

export interface MySQLTable {
  name: string;
  storageEngine: string;
  characterSet?: string;
  collation?: string;
  comment?: string;
  columns: MySQLColumn[];
  primaryKey?: MySQLPrimaryKey;
  indexes: MySQLIndex[];
  foreignKeys: MySQLForeignKey[];
}

export interface MySQLColumn {
  name: string;
  position: number;
  type: MySQLType;
  nullable: boolean;
  defaultExpression?: string;
  generatedExpression?: string;
  autoIncrement: boolean;
  characterSet?: string;
  collation?: string;
  comment?: string;
}

export interface MySQLType {
  name: string;
  nativeType: string;
  length?: number;
  precision?: number;
  scale?: number;
  unsigned: boolean;
  enumValues?: string[];
  setValues?: string[];
}

export interface MySQLPrimaryKey {
  name?: string;
  columns: string[];
}

export interface MySQLIndex {
  name: string;
  unique: boolean;
  kind: "btree" | "hash" | "fulltext" | "spatial" | "other";
  parts: MySQLIndexPart[];
  visible: boolean;
}

export interface MySQLIndexPart {
  column?: string;
  expression?: string;
  prefixLength?: number;
  descending: boolean;
}

export interface MySQLForeignKey {
  name?: string;
  columns: MySQLForeignKeyColumn[];
  referencedTable: MySQLTableReference;
  onUpdate: MySQLReferentialAction;
  onDelete: MySQLReferentialAction;
}

export interface MySQLForeignKeyColumn {
  column: string;
  referencedColumn: string;
}

export interface MySQLTableReference {
  database: string;
  table: string;
}

export type MySQLReferentialAction =
  | "no-action"
  | "restrict"
  | "cascade"
  | "set-null";

export interface MySQLRelationship {
  fromTable: MySQLTableReference;
  fromColumns: string[];
  toTable: MySQLTableReference;
  toColumns: string[];
  foreignKey?: string;
  cardinality: "many-to-one" | "one-to-one";
  optional: boolean;
}
