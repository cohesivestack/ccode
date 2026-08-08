export interface PostgreSQLInspection {
  engine: "postgresql";
  database: PostgreSQLDatabase;
}

export interface PostgreSQLDatabase {
  name: string;
  schemas: PostgreSQLSchema[];
  relationships: PostgreSQLRelationship[];
}

export interface PostgreSQLSchema {
  name: string;
  enumTypes: PostgreSQLEnumType[];
  tables: PostgreSQLTable[];
}

export interface PostgreSQLEnumType {
  name: string;
  values: string[];
}

export interface PostgreSQLTable {
  name: string;
  comment?: string;
  columns: PostgreSQLColumn[];
  primaryKey?: PostgreSQLPrimaryKey;
  indexes: PostgreSQLIndex[];
  foreignKeys: PostgreSQLForeignKey[];
}

export interface PostgreSQLColumn {
  name: string;
  position: number;
  type: PostgreSQLType;
  nullable: boolean;
  defaultExpression?: string;
  generatedExpression?: string;
  identity?: "always" | "by-default";
  comment?: string;
}

export interface PostgreSQLType {
  name: string;
  nativeType: string;
  schema?: string;
  length?: number;
  precision?: number;
  scale?: number;
  arrayDimensions?: number;
}

export interface PostgreSQLPrimaryKey {
  name?: string;
  columns: string[];
  deferrable: boolean;
  initiallyDeferred: boolean;
}

export interface PostgreSQLIndex {
  name: string;
  unique: boolean;
  method?: string;
  parts: PostgreSQLIndexPart[];
  includedColumns: string[];
  predicate?: string;
}

export interface PostgreSQLIndexPart {
  column?: string;
  expression?: string;
  descending: boolean;
}

export interface PostgreSQLForeignKey {
  name?: string;
  columns: PostgreSQLForeignKeyColumn[];
  referencedTable: PostgreSQLTableReference;
  onUpdate: PostgreSQLReferentialAction;
  onDelete: PostgreSQLReferentialAction;
  deferrable: boolean;
  initiallyDeferred: boolean;
}

export interface PostgreSQLForeignKeyColumn {
  column: string;
  referencedColumn: string;
}

export interface PostgreSQLTableReference {
  schema: string;
  table: string;
}

export type PostgreSQLReferentialAction =
  | "no-action"
  | "restrict"
  | "cascade"
  | "set-null"
  | "set-default";

export interface PostgreSQLRelationship {
  fromTable: PostgreSQLTableReference;
  fromColumns: string[];
  toTable: PostgreSQLTableReference;
  toColumns: string[];
  foreignKey?: string;
  cardinality: "many-to-one" | "one-to-one";
  optional: boolean;
}
