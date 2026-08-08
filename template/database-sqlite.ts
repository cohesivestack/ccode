export interface SQLiteInspection {
  engine: "sqlite";
  databases: SQLiteDatabase[];
}

export interface SQLiteDatabase {
  name: string;
  tables: SQLiteTable[];
  relationships: SQLiteRelationship[];
}

export interface SQLiteTable {
  name: string;
  strict: boolean;
  withoutRowID: boolean;
  columns: SQLiteColumn[];
  primaryKey?: SQLitePrimaryKey;
  indexes: SQLiteIndex[];
  foreignKeys: SQLiteForeignKey[];
}

export interface SQLiteColumn {
  name: string;
  position: number;
  type: SQLiteType;
  nullable: boolean;
  defaultExpression?: string;
  generatedExpression?: string;
}

export interface SQLiteType {
  declaredType: string;
  affinity: "integer" | "real" | "text" | "blob" | "numeric";
}

export interface SQLitePrimaryKey {
  columns: string[];
}

export interface SQLiteIndex {
  name: string;
  unique: boolean;
  parts: SQLiteIndexPart[];
  predicate?: string;
}

export interface SQLiteIndexPart {
  column?: string;
  expression?: string;
  descending: boolean;
}

export interface SQLiteForeignKey {
  columns: SQLiteForeignKeyColumn[];
  referencedTable: SQLiteTableReference;
  onUpdate: SQLiteReferentialAction;
  onDelete: SQLiteReferentialAction;
  deferrable: boolean;
  initiallyDeferred: boolean;
}

export interface SQLiteForeignKeyColumn {
  column: string;
  referencedColumn: string;
}

export interface SQLiteTableReference {
  table: string;
}

export type SQLiteReferentialAction =
  | "no-action"
  | "restrict"
  | "cascade"
  | "set-null"
  | "set-default";

export interface SQLiteRelationship {
  fromTable: SQLiteTableReference;
  fromColumns: string[];
  toTable: SQLiteTableReference;
  toColumns: string[];
  cardinality: "many-to-one" | "one-to-one";
  optional: boolean;
}
