package ccode

import (
	"strings"

	"ariga.io/atlas/sql/schema"
	atlassqlite "ariga.io/atlas/sql/sqlite"
)

type SQLiteInspection struct {
	Engine    DatabaseEngine   `json:"engine"`
	Databases []SQLiteDatabase `json:"databases"`
}

type SQLiteDatabase struct {
	Name          string               `json:"name"`
	Tables        []SQLiteTable        `json:"tables"`
	Relationships []SQLiteRelationship `json:"relationships"`
}

type SQLiteTable struct {
	Name         string             `json:"name"`
	Strict       bool               `json:"strict"`
	WithoutRowID bool               `json:"withoutRowID"`
	Columns      []SQLiteColumn     `json:"columns"`
	PrimaryKey   *SQLitePrimaryKey  `json:"primaryKey,omitempty"`
	Indexes      []SQLiteIndex      `json:"indexes"`
	ForeignKeys  []SQLiteForeignKey `json:"foreignKeys"`
}

type SQLiteColumn struct {
	Name                string     `json:"name"`
	Position            int        `json:"position"`
	Type                SQLiteType `json:"type"`
	Nullable            bool       `json:"nullable"`
	DefaultExpression   string     `json:"defaultExpression,omitempty"`
	GeneratedExpression string     `json:"generatedExpression,omitempty"`
}

type SQLiteType struct {
	DeclaredType string         `json:"declaredType"`
	Affinity     SQLiteAffinity `json:"affinity"`
}

type SQLiteAffinity string

const (
	SQLiteAffinityInteger SQLiteAffinity = "integer"
	SQLiteAffinityReal    SQLiteAffinity = "real"
	SQLiteAffinityText    SQLiteAffinity = "text"
	SQLiteAffinityBlob    SQLiteAffinity = "blob"
	SQLiteAffinityNumeric SQLiteAffinity = "numeric"
)

type SQLitePrimaryKey struct {
	Columns []string `json:"columns"`
}

type SQLiteIndex struct {
	Name      string            `json:"name"`
	Unique    bool              `json:"unique"`
	Parts     []SQLiteIndexPart `json:"parts"`
	Predicate string            `json:"predicate,omitempty"`
}

type SQLiteIndexPart struct {
	Column     string `json:"column,omitempty"`
	Expression string `json:"expression,omitempty"`
	Descending bool   `json:"descending"`
}

type SQLiteForeignKey struct {
	Columns           []SQLiteForeignKeyColumn `json:"columns"`
	ReferencedTable   SQLiteTableReference     `json:"referencedTable"`
	OnUpdate          SQLiteReferentialAction  `json:"onUpdate"`
	OnDelete          SQLiteReferentialAction  `json:"onDelete"`
	Deferrable        bool                     `json:"deferrable"`
	InitiallyDeferred bool                     `json:"initiallyDeferred"`
}

type SQLiteForeignKeyColumn struct {
	Column           string `json:"column"`
	ReferencedColumn string `json:"referencedColumn"`
}

type SQLiteTableReference struct {
	Table string `json:"table"`
}

type SQLiteReferentialAction string

const (
	SQLiteNoAction   SQLiteReferentialAction = "no-action"
	SQLiteRestrict   SQLiteReferentialAction = "restrict"
	SQLiteCascade    SQLiteReferentialAction = "cascade"
	SQLiteSetNull    SQLiteReferentialAction = "set-null"
	SQLiteSetDefault SQLiteReferentialAction = "set-default"
)

type SQLiteRelationship struct {
	FromTable   SQLiteTableReference `json:"fromTable"`
	FromColumns []string             `json:"fromColumns"`
	ToTable     SQLiteTableReference `json:"toTable"`
	ToColumns   []string             `json:"toColumns"`
	Cardinality string               `json:"cardinality"`
	Optional    bool                 `json:"optional"`
}

func (SQLiteInspection) databaseInspection() {}

func convertSQLiteRealm(realm *schema.Realm) *SQLiteInspection {
	inspection := &SQLiteInspection{
		Engine:    DatabaseEngineSQLite,
		Databases: make([]SQLiteDatabase, 0, len(realm.Schemas)),
	}
	for _, atlasDatabase := range realm.Schemas {
		if atlasDatabase == nil {
			continue
		}
		database := SQLiteDatabase{
			Name:          atlasDatabase.Name,
			Tables:        make([]SQLiteTable, 0, len(atlasDatabase.Tables)),
			Relationships: make([]SQLiteRelationship, 0),
		}
		for _, atlasTable := range atlasDatabase.Tables {
			if atlasTable == nil {
				continue
			}
			database.Tables = append(database.Tables, convertSQLiteTable(atlasTable))
			database.Relationships = append(
				database.Relationships,
				convertSQLiteRelationships(atlasDatabase, atlasTable)...,
			)
		}
		inspection.Databases = append(inspection.Databases, database)
	}
	return inspection
}

func convertSQLiteTable(table *schema.Table) SQLiteTable {
	_, strict := databaseAttribute[*atlassqlite.Strict](table.Attrs)
	_, withoutRowID := databaseAttribute[*atlassqlite.WithoutRowID](table.Attrs)
	converted := SQLiteTable{
		Name:         table.Name,
		Strict:       strict,
		WithoutRowID: withoutRowID,
		Columns:      make([]SQLiteColumn, 0, len(table.Columns)),
		Indexes:      make([]SQLiteIndex, 0, len(table.Indexes)),
		ForeignKeys:  make([]SQLiteForeignKey, 0, len(table.ForeignKeys)),
	}
	for position, column := range table.Columns {
		if column != nil {
			converted.Columns = append(converted.Columns, convertSQLiteColumn(column, position+1))
		}
	}
	if table.PrimaryKey != nil {
		converted.PrimaryKey = &SQLitePrimaryKey{
			Columns: databaseIndexColumnNames(table.PrimaryKey),
		}
	}
	for _, index := range table.Indexes {
		if index != nil {
			converted.Indexes = append(converted.Indexes, convertSQLiteIndex(index))
		}
	}
	for _, foreignKey := range table.ForeignKeys {
		if foreignKey != nil {
			converted.ForeignKeys = append(converted.ForeignKeys, convertSQLiteForeignKey(foreignKey))
		}
	}
	return converted
}

func convertSQLiteColumn(column *schema.Column, position int) SQLiteColumn {
	converted := SQLiteColumn{
		Name:                column.Name,
		Position:            position,
		Type:                convertSQLiteType(column.Type),
		DefaultExpression:   databaseExpression(column.Default),
		GeneratedExpression: databaseGeneratedExpression(column.Attrs),
	}
	if column.Type != nil {
		converted.Nullable = column.Type.Null
	}
	return converted
}

func convertSQLiteType(columnType *schema.ColumnType) SQLiteType {
	if columnType == nil {
		return SQLiteType{Affinity: SQLiteAffinityBlob}
	}
	declaredType := columnType.Raw
	if declaredType == "" {
		declaredType = sqliteTypeName(columnType.Type)
	}
	return SQLiteType{
		DeclaredType: declaredType,
		Affinity:     sqliteTypeAffinity(declaredType),
	}
}

func sqliteTypeName(atlasType schema.Type) string {
	switch atlasType := atlasType.(type) {
	case *atlassqlite.UserDefinedType:
		return atlasType.T
	case *schema.StringType:
		return atlasType.T
	case *schema.BinaryType:
		return atlasType.T
	case *schema.IntegerType:
		return atlasType.T
	case *schema.DecimalType:
		return atlasType.T
	case *schema.FloatType:
		return atlasType.T
	case *schema.TimeType:
		return atlasType.T
	case *schema.BoolType:
		return atlasType.T
	case *schema.JSONType:
		return atlasType.T
	case *schema.SpatialType:
		return atlasType.T
	case *schema.UUIDType:
		return atlasType.T
	case *schema.UnsupportedType:
		return atlasType.T
	default:
		return ""
	}
}

func sqliteTypeAffinity(declaredType string) SQLiteAffinity {
	normalized := strings.ToUpper(declaredType)
	switch {
	case strings.Contains(normalized, "INT"):
		return SQLiteAffinityInteger
	case strings.Contains(normalized, "CHAR"), strings.Contains(normalized, "CLOB"), strings.Contains(normalized, "TEXT"):
		return SQLiteAffinityText
	case normalized == "", strings.Contains(normalized, "BLOB"):
		return SQLiteAffinityBlob
	case strings.Contains(normalized, "REAL"), strings.Contains(normalized, "FLOA"), strings.Contains(normalized, "DOUB"):
		return SQLiteAffinityReal
	default:
		return SQLiteAffinityNumeric
	}
}

func convertSQLiteIndex(index *schema.Index) SQLiteIndex {
	converted := SQLiteIndex{
		Name:   index.Name,
		Unique: index.Unique,
		Parts:  make([]SQLiteIndexPart, 0, len(index.Parts)),
	}
	if predicate, ok := databaseAttribute[*atlassqlite.IndexPredicate](index.Attrs); ok {
		converted.Predicate = predicate.P
	}
	for _, part := range index.Parts {
		if part == nil {
			continue
		}
		convertedPart := SQLiteIndexPart{
			Expression: databaseIndexPartExpression(part),
			Descending: part.Desc,
		}
		if part.C != nil {
			convertedPart.Column = part.C.Name
		}
		converted.Parts = append(converted.Parts, convertedPart)
	}
	return converted
}

func convertSQLiteForeignKey(foreignKey *schema.ForeignKey) SQLiteForeignKey {
	converted := SQLiteForeignKey{
		Columns:           make([]SQLiteForeignKeyColumn, 0, len(foreignKey.Columns)),
		ReferencedTable:   sqliteTableReference(foreignKey.RefTable),
		OnUpdate:          SQLiteReferentialAction(databaseReferenceAction(foreignKey.OnUpdate)),
		OnDelete:          SQLiteReferentialAction(databaseReferenceAction(foreignKey.OnDelete)),
		Deferrable:        false,
		InitiallyDeferred: false,
	}
	for position, column := range foreignKey.Columns {
		if column == nil || position >= len(foreignKey.RefColumns) || foreignKey.RefColumns[position] == nil {
			continue
		}
		converted.Columns = append(converted.Columns, SQLiteForeignKeyColumn{
			Column:           column.Name,
			ReferencedColumn: foreignKey.RefColumns[position].Name,
		})
	}
	return converted
}

func convertSQLiteRelationships(database *schema.Schema, table *schema.Table) []SQLiteRelationship {
	relationships := make([]SQLiteRelationship, 0, len(table.ForeignKeys))
	for _, foreignKey := range table.ForeignKeys {
		if foreignKey == nil || foreignKey.RefTable == nil || foreignKey.RefTable.Schema == nil || foreignKey.RefTable.Schema.Name != database.Name {
			continue
		}
		if _, ok := database.Table(foreignKey.RefTable.Name); !ok {
			continue
		}
		relationships = append(relationships, SQLiteRelationship{
			FromTable:   sqliteTableReference(table),
			FromColumns: databaseColumnNames(foreignKey.Columns),
			ToTable:     sqliteTableReference(foreignKey.RefTable),
			ToColumns:   databaseColumnNames(foreignKey.RefColumns),
			Cardinality: databaseForeignKeyCardinality(table, foreignKey.Columns, sqliteIndexIsPartial),
			Optional:    databaseForeignKeyOptional(foreignKey.Columns),
		})
	}
	return relationships
}

func sqliteTableReference(table *schema.Table) SQLiteTableReference {
	if table == nil {
		return SQLiteTableReference{}
	}
	return SQLiteTableReference{Table: table.Name}
}

func sqliteIndexIsPartial(index *schema.Index) bool {
	_, ok := databaseAttribute[*atlassqlite.IndexPredicate](index.Attrs)
	return ok
}
