export type ConnectionURL =
  | `postgres://${string}`
  | `postgresql://${string}`;

export interface Inspection {
  engine: "postgresql";
  database: Database;
}

export interface Database {
  name: string;
  schemas: Schema[];
  relationships: Relationship[];
}

export interface Schema {
  name: string;
  enumTypes: EnumType[];
  tables: Table[];
}

export interface EnumType {
  name: string;
  values: string[];
}

export interface Table {
  name: string;
  comment?: string;
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
  identity?: "always" | "by-default";
  comment?: string;
}

export interface Type {
  name: string;
  nativeType: string;
  schema?: string;
  length?: number;
  precision?: number;
  scale?: number;
  arrayDimensions?: number;
}

export interface PrimaryKey {
  name?: string;
  columns: string[];
  deferrable: boolean;
  initiallyDeferred: boolean;
}

export interface Index {
  name: string;
  unique: boolean;
  method?: string;
  parts: IndexPart[];
  includedColumns: string[];
  predicate?: string;
}

export interface IndexPart {
  column?: string;
  expression?: string;
  descending: boolean;
}

export interface ForeignKey {
  name?: string;
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
  schema: string;
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
  foreignKey?: string;
  cardinality: "many-to-one" | "one-to-one";
  optional: boolean;
}

export function isInspection(
  inspection: { engine: string },
): inspection is Inspection {
  return inspection.engine === "postgresql";
}

