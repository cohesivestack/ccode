export type ConnectionURL =
  | `maria://${string}`
  | `mariadb://${string}`;

export interface Inspection {
  engine: "mariadb";
  databases: Database[];
  relationships: Relationship[];
}

export interface Database {
  name: string;
  characterSet?: string;
  collation?: string;
  tables: Table[];
}

export interface Table {
  name: string;
  storageEngine: string;
  characterSet?: string;
  collation?: string;
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
  autoIncrement: boolean;
  characterSet?: string;
  collation?: string;
  comment?: string;
}

export interface Type {
  name: string;
  nativeType: string;
  length?: number;
  precision?: number;
  scale?: number;
  unsigned: boolean;
  enumValues?: string[];
  setValues?: string[];
}

export interface PrimaryKey {
  name?: string;
  columns: string[];
}

export interface Index {
  name: string;
  unique: boolean;
  kind: "btree" | "hash" | "fulltext" | "spatial" | "other";
  parts: IndexPart[];
  ignored: boolean;
}

export interface IndexPart {
  column?: string;
  expression?: string;
  prefixLength?: number;
  descending: boolean;
}

export interface ForeignKey {
  name?: string;
  columns: ForeignKeyColumn[];
  referencedTable: TableReference;
  onUpdate: ReferentialAction;
  onDelete: ReferentialAction;
}

export interface ForeignKeyColumn {
  column: string;
  referencedColumn: string;
}

export interface TableReference {
  database: string;
  table: string;
}

export type ReferentialAction =
  | "no-action"
  | "restrict"
  | "cascade"
  | "set-null";

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
  return inspection.engine === "mariadb";
}

