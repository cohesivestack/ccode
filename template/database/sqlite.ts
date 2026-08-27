export type ConnectionURL = `sqlite://${string}`;

export interface Inspection {
  engine: "sqlite";
  databases: Database[];
}

export interface Database {
  name: string;
  tables: Table[];
  relationships: Relationship[];
}

export interface Table {
  name: string;
  strict: boolean;
  withoutRowID: boolean;
  columns: Column[];
  primaryKey?: PrimaryKey;
  indexes: Index[];
  foreignKeys: ForeignKey[];
}

export interface Column {
  name: string;
  position: number;
  type: Type;
  nullable: boolean;
  defaultExpression?: string;
  generatedExpression?: string;
}

export interface Type {
  declaredType: string;
  affinity: "integer" | "real" | "text" | "blob" | "numeric";
}

export interface PrimaryKey {
  columns: string[];
}

export interface Index {
  name: string;
  unique: boolean;
  parts: IndexPart[];
  predicate?: string;
}

export interface IndexPart {
  column?: string;
  expression?: string;
  descending: boolean;
}

export interface ForeignKey {
  columns: ForeignKeyColumn[];
  referencedTable: TableReference;
  onUpdate: ReferentialAction;
  onDelete: ReferentialAction;
  deferrable: boolean;
  initiallyDeferred: boolean;
}

export interface ForeignKeyColumn {
  column: string;
  referencedColumn: string;
}

export interface TableReference {
  table: string;
}

export type ReferentialAction =
  | "no-action"
  | "restrict"
  | "cascade"
  | "set-null"
  | "set-default";

export interface Relationship {
  fromTable: TableReference;
  fromColumns: string[];
  toTable: TableReference;
  toColumns: string[];
  cardinality: "many-to-one" | "one-to-one";
  optional: boolean;
}

export function isInspection(
  inspection: { engine: string },
): inspection is Inspection {
  return inspection.engine === "sqlite";
}

