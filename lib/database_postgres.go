package ccode

import (
	"strings"

	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
)

type PostgreSQLInspection struct {
	Engine   DatabaseEngine     `json:"engine"`
	Database PostgreSQLDatabase `json:"database"`
}

type PostgreSQLDatabase struct {
	Name          string                   `json:"name"`
	Schemas       []PostgreSQLSchema       `json:"schemas"`
	Relationships []PostgreSQLRelationship `json:"relationships"`
}

type PostgreSQLSchema struct {
	Name      string               `json:"name"`
	EnumTypes []PostgreSQLEnumType `json:"enumTypes"`
	Tables    []PostgreSQLTable    `json:"tables"`
}

type PostgreSQLEnumType struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type PostgreSQLTable struct {
	Name        string                 `json:"name"`
	Comment     string                 `json:"comment,omitempty"`
	Columns     []PostgreSQLColumn     `json:"columns"`
	PrimaryKey  *PostgreSQLPrimaryKey  `json:"primaryKey,omitempty"`
	Indexes     []PostgreSQLIndex      `json:"indexes"`
	ForeignKeys []PostgreSQLForeignKey `json:"foreignKeys"`
}

type PostgreSQLColumn struct {
	Name                string         `json:"name"`
	Position            int            `json:"position"`
	Type                PostgreSQLType `json:"type"`
	Nullable            bool           `json:"nullable"`
	DefaultExpression   string         `json:"defaultExpression,omitempty"`
	GeneratedExpression string         `json:"generatedExpression,omitempty"`
	Identity            string         `json:"identity,omitempty"`
	Comment             string         `json:"comment,omitempty"`
}

type PostgreSQLType struct {
	Name            string `json:"name"`
	NativeType      string `json:"nativeType"`
	Schema          string `json:"schema,omitempty"`
	Length          int    `json:"length,omitempty"`
	Precision       int    `json:"precision,omitempty"`
	Scale           int    `json:"scale,omitempty"`
	ArrayDimensions int    `json:"arrayDimensions,omitempty"`
}

type PostgreSQLPrimaryKey struct {
	Name              string   `json:"name,omitempty"`
	Columns           []string `json:"columns"`
	Deferrable        bool     `json:"deferrable"`
	InitiallyDeferred bool     `json:"initiallyDeferred"`
}

type PostgreSQLIndex struct {
	Name            string                `json:"name"`
	Unique          bool                  `json:"unique"`
	Method          string                `json:"method,omitempty"`
	Parts           []PostgreSQLIndexPart `json:"parts"`
	IncludedColumns []string              `json:"includedColumns"`
	Predicate       string                `json:"predicate,omitempty"`
}

type PostgreSQLIndexPart struct {
	Column     string `json:"column,omitempty"`
	Expression string `json:"expression,omitempty"`
	Descending bool   `json:"descending"`
}

type PostgreSQLForeignKey struct {
	Name              string                       `json:"name,omitempty"`
	Columns           []PostgreSQLForeignKeyColumn `json:"columns"`
	ReferencedTable   PostgreSQLTableReference     `json:"referencedTable"`
	OnUpdate          PostgreSQLReferentialAction  `json:"onUpdate"`
	OnDelete          PostgreSQLReferentialAction  `json:"onDelete"`
	Deferrable        bool                         `json:"deferrable"`
	InitiallyDeferred bool                         `json:"initiallyDeferred"`
}

type PostgreSQLForeignKeyColumn struct {
	Column           string `json:"column"`
	ReferencedColumn string `json:"referencedColumn"`
}

type PostgreSQLTableReference struct {
	Schema string `json:"schema"`
	Table  string `json:"table"`
}

type PostgreSQLReferentialAction string

const (
	PostgreSQLNoAction   PostgreSQLReferentialAction = "no-action"
	PostgreSQLRestrict   PostgreSQLReferentialAction = "restrict"
	PostgreSQLCascade    PostgreSQLReferentialAction = "cascade"
	PostgreSQLSetNull    PostgreSQLReferentialAction = "set-null"
	PostgreSQLSetDefault PostgreSQLReferentialAction = "set-default"
)

type PostgreSQLRelationship struct {
	FromTable   PostgreSQLTableReference `json:"fromTable"`
	FromColumns []string                 `json:"fromColumns"`
	ToTable     PostgreSQLTableReference `json:"toTable"`
	ToColumns   []string                 `json:"toColumns"`
	ForeignKey  string                   `json:"foreignKey,omitempty"`
	Cardinality string                   `json:"cardinality"`
	Optional    bool                     `json:"optional"`
}

func (PostgreSQLInspection) databaseInspection() {}

func convertPostgreSQLRealm(realm *schema.Realm, databaseName string) *PostgreSQLInspection {
	inspection := &PostgreSQLInspection{
		Engine: DatabaseEnginePostgreSQL,
		Database: PostgreSQLDatabase{
			Name:          databaseName,
			Schemas:       make([]PostgreSQLSchema, 0, len(realm.Schemas)),
			Relationships: make([]PostgreSQLRelationship, 0),
		},
	}

	for _, atlasSchema := range realm.Schemas {
		if atlasSchema == nil {
			continue
		}
		convertedSchema := PostgreSQLSchema{
			Name:      atlasSchema.Name,
			EnumTypes: convertPostgreSQLEnumTypes(atlasSchema),
			Tables:    make([]PostgreSQLTable, 0, len(atlasSchema.Tables)),
		}
		for _, atlasTable := range atlasSchema.Tables {
			if atlasTable == nil {
				continue
			}
			convertedSchema.Tables = append(convertedSchema.Tables, convertPostgreSQLTable(atlasTable))
			inspection.Database.Relationships = append(
				inspection.Database.Relationships,
				convertPostgreSQLRelationships(realm, atlasTable)...,
			)
		}
		inspection.Database.Schemas = append(inspection.Database.Schemas, convertedSchema)
	}
	return inspection
}

func convertPostgreSQLEnumTypes(atlasSchema *schema.Schema) []PostgreSQLEnumType {
	enums := make([]PostgreSQLEnumType, 0)
	for _, object := range atlasSchema.Objects {
		enum, ok := object.(*schema.EnumType)
		if !ok {
			continue
		}
		enums = append(enums, PostgreSQLEnumType{
			Name:   enum.T,
			Values: append([]string(nil), enum.Values...),
		})
	}
	return enums
}

func convertPostgreSQLTable(table *schema.Table) PostgreSQLTable {
	converted := PostgreSQLTable{
		Name:        table.Name,
		Comment:     databaseComment(table.Attrs),
		Columns:     make([]PostgreSQLColumn, 0, len(table.Columns)),
		Indexes:     make([]PostgreSQLIndex, 0, len(table.Indexes)),
		ForeignKeys: make([]PostgreSQLForeignKey, 0, len(table.ForeignKeys)),
	}
	for position, column := range table.Columns {
		if column != nil {
			converted.Columns = append(converted.Columns, convertPostgreSQLColumn(column, position+1))
		}
	}
	if table.PrimaryKey != nil {
		converted.PrimaryKey = &PostgreSQLPrimaryKey{
			Name:              table.PrimaryKey.Name,
			Columns:           databaseIndexColumnNames(table.PrimaryKey),
			Deferrable:        false,
			InitiallyDeferred: false,
		}
	}
	for _, index := range table.Indexes {
		if index != nil {
			converted.Indexes = append(converted.Indexes, convertPostgreSQLIndex(index))
		}
	}
	for _, foreignKey := range table.ForeignKeys {
		if foreignKey != nil {
			converted.ForeignKeys = append(converted.ForeignKeys, convertPostgreSQLForeignKey(foreignKey))
		}
	}
	return converted
}

func convertPostgreSQLColumn(column *schema.Column, position int) PostgreSQLColumn {
	converted := PostgreSQLColumn{
		Name:                column.Name,
		Position:            position,
		Type:                convertPostgreSQLType(column.Type),
		DefaultExpression:   databaseExpression(column.Default),
		GeneratedExpression: databaseGeneratedExpression(column.Attrs),
		Comment:             databaseComment(column.Attrs),
	}
	if column.Type != nil {
		converted.Nullable = column.Type.Null
	}
	if identity, ok := databaseAttribute[*postgres.Identity](column.Attrs); ok {
		converted.Identity = strings.ToLower(strings.ReplaceAll(identity.Generation, " ", "-"))
	}
	return converted
}

func convertPostgreSQLType(columnType *schema.ColumnType) PostgreSQLType {
	if columnType == nil {
		return PostgreSQLType{}
	}
	converted := PostgreSQLType{NativeType: columnType.Raw}
	if converted.NativeType == "" && columnType.Type != nil {
		if nativeType, err := postgres.FormatType(columnType.Type); err == nil {
			converted.NativeType = nativeType
		}
	}
	populatePostgreSQLType(&converted, columnType.Type)
	if converted.Name == "" {
		converted.Name = converted.NativeType
	}
	return converted
}

func populatePostgreSQLType(converted *PostgreSQLType, atlasType schema.Type) {
	switch atlasType := atlasType.(type) {
	case *postgres.ArrayType:
		populatePostgreSQLType(converted, atlasType.Type)
		converted.ArrayDimensions += strings.Count(atlasType.T, "[]")
		if converted.ArrayDimensions == 0 {
			converted.ArrayDimensions = 1
		}
	case *schema.EnumType:
		converted.Name = atlasType.T
		if atlasType.Schema != nil {
			converted.Schema = atlasType.Schema.Name
		}
	case *postgres.DomainType:
		converted.Name = atlasType.T
		if atlasType.Schema != nil {
			converted.Schema = atlasType.Schema.Name
		}
	case *postgres.UserDefinedType:
		converted.Name = atlasType.T
	case *postgres.SerialType:
		converted.Name = atlasType.T
		converted.Precision = atlasType.Precision
	case *postgres.BitType:
		converted.Name = atlasType.T
		converted.Length = int(atlasType.Len)
	case *postgres.IntervalType:
		converted.Name = atlasType.T
		if atlasType.Precision != nil {
			converted.Precision = *atlasType.Precision
		}
	case *postgres.NetworkType:
		converted.Name = atlasType.T
		converted.Length = int(atlasType.Len)
	case *postgres.CurrencyType:
		converted.Name = atlasType.T
	case *postgres.RangeType:
		converted.Name = atlasType.T
	case *postgres.TextSearchType:
		converted.Name = atlasType.T
	case *postgres.OIDType:
		converted.Name = atlasType.T
	case *postgres.XMLType:
		converted.Name = atlasType.T
	case *schema.StringType:
		converted.Name = atlasType.T
		converted.Length = atlasType.Size
	case *schema.BinaryType:
		converted.Name = atlasType.T
		if atlasType.Size != nil {
			converted.Length = *atlasType.Size
		}
	case *schema.IntegerType:
		converted.Name = atlasType.T
	case *schema.DecimalType:
		converted.Name = atlasType.T
		converted.Precision = atlasType.Precision
		converted.Scale = atlasType.Scale
	case *schema.FloatType:
		converted.Name = atlasType.T
		converted.Precision = atlasType.Precision
	case *schema.TimeType:
		converted.Name = atlasType.T
		if atlasType.Precision != nil {
			converted.Precision = *atlasType.Precision
		}
		if atlasType.Scale != nil {
			converted.Scale = *atlasType.Scale
		}
	case *schema.BoolType:
		converted.Name = atlasType.T
	case *schema.JSONType:
		converted.Name = atlasType.T
	case *schema.SpatialType:
		converted.Name = atlasType.T
	case *schema.UUIDType:
		converted.Name = atlasType.T
	case *schema.UnsupportedType:
		converted.Name = atlasType.T
	}
}

func convertPostgreSQLIndex(index *schema.Index) PostgreSQLIndex {
	converted := PostgreSQLIndex{
		Name:            index.Name,
		Unique:          index.Unique,
		Parts:           make([]PostgreSQLIndexPart, 0, len(index.Parts)),
		IncludedColumns: make([]string, 0),
	}
	if indexType, ok := databaseAttribute[*postgres.IndexType](index.Attrs); ok {
		converted.Method = strings.ToLower(indexType.T)
	}
	if predicate, ok := databaseAttribute[*postgres.IndexPredicate](index.Attrs); ok {
		converted.Predicate = predicate.P
	}
	if include, ok := databaseAttribute[*postgres.IndexInclude](index.Attrs); ok {
		converted.IncludedColumns = databaseColumnNames(include.Columns)
	}
	for _, part := range index.Parts {
		if part == nil {
			continue
		}
		convertedPart := PostgreSQLIndexPart{
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

func convertPostgreSQLForeignKey(foreignKey *schema.ForeignKey) PostgreSQLForeignKey {
	converted := PostgreSQLForeignKey{
		Name:              foreignKey.Symbol,
		Columns:           make([]PostgreSQLForeignKeyColumn, 0, len(foreignKey.Columns)),
		ReferencedTable:   postgreSQLTableReference(foreignKey.RefTable),
		OnUpdate:          PostgreSQLReferentialAction(databaseReferenceAction(foreignKey.OnUpdate)),
		OnDelete:          PostgreSQLReferentialAction(databaseReferenceAction(foreignKey.OnDelete)),
		Deferrable:        false,
		InitiallyDeferred: false,
	}
	for position, column := range foreignKey.Columns {
		if column == nil || position >= len(foreignKey.RefColumns) || foreignKey.RefColumns[position] == nil {
			continue
		}
		converted.Columns = append(converted.Columns, PostgreSQLForeignKeyColumn{
			Column:           column.Name,
			ReferencedColumn: foreignKey.RefColumns[position].Name,
		})
	}
	return converted
}

func convertPostgreSQLRelationships(realm *schema.Realm, table *schema.Table) []PostgreSQLRelationship {
	relationships := make([]PostgreSQLRelationship, 0, len(table.ForeignKeys))
	for _, foreignKey := range table.ForeignKeys {
		if foreignKey == nil || !databaseTableInRealm(realm, foreignKey.RefTable) {
			continue
		}
		relationships = append(relationships, PostgreSQLRelationship{
			FromTable:   postgreSQLTableReference(table),
			FromColumns: databaseColumnNames(foreignKey.Columns),
			ToTable:     postgreSQLTableReference(foreignKey.RefTable),
			ToColumns:   databaseColumnNames(foreignKey.RefColumns),
			ForeignKey:  foreignKey.Symbol,
			Cardinality: databaseForeignKeyCardinality(table, foreignKey.Columns, postgreSQLIndexIsPartial),
			Optional:    databaseForeignKeyOptional(foreignKey.Columns),
		})
	}
	return relationships
}

func postgreSQLTableReference(table *schema.Table) PostgreSQLTableReference {
	reference := PostgreSQLTableReference{}
	if table == nil {
		return reference
	}
	reference.Table = table.Name
	if table.Schema != nil {
		reference.Schema = table.Schema.Name
	}
	return reference
}

func postgreSQLIndexIsPartial(index *schema.Index) bool {
	_, ok := databaseAttribute[*postgres.IndexPredicate](index.Attrs)
	return ok
}
