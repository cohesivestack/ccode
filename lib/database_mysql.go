package ccode

import (
	"strings"

	atlasmysql "ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/schema"
)

type MySQLInspection struct {
	Engine        DatabaseEngine      `json:"engine"`
	Databases     []MySQLDatabase     `json:"databases"`
	Relationships []MySQLRelationship `json:"relationships"`
}

type MySQLDatabase struct {
	Name         string       `json:"name"`
	CharacterSet string       `json:"characterSet,omitempty"`
	Collation    string       `json:"collation,omitempty"`
	Tables       []MySQLTable `json:"tables"`
}

type MySQLTable struct {
	Name          string            `json:"name"`
	StorageEngine string            `json:"storageEngine"`
	CharacterSet  string            `json:"characterSet,omitempty"`
	Collation     string            `json:"collation,omitempty"`
	Comment       string            `json:"comment,omitempty"`
	Columns       []MySQLColumn     `json:"columns"`
	PrimaryKey    *MySQLPrimaryKey  `json:"primaryKey,omitempty"`
	Indexes       []MySQLIndex      `json:"indexes"`
	ForeignKeys   []MySQLForeignKey `json:"foreignKeys"`
}

type MySQLColumn struct {
	Name                string    `json:"name"`
	Position            int       `json:"position"`
	Type                MySQLType `json:"type"`
	Nullable            bool      `json:"nullable"`
	DefaultExpression   string    `json:"defaultExpression,omitempty"`
	GeneratedExpression string    `json:"generatedExpression,omitempty"`
	AutoIncrement       bool      `json:"autoIncrement"`
	CharacterSet        string    `json:"characterSet,omitempty"`
	Collation           string    `json:"collation,omitempty"`
	Comment             string    `json:"comment,omitempty"`
}

type MySQLType struct {
	Name       string   `json:"name"`
	NativeType string   `json:"nativeType"`
	Length     int      `json:"length,omitempty"`
	Precision  int      `json:"precision,omitempty"`
	Scale      int      `json:"scale,omitempty"`
	Unsigned   bool     `json:"unsigned"`
	EnumValues []string `json:"enumValues,omitempty"`
	SetValues  []string `json:"setValues,omitempty"`
}

type MySQLPrimaryKey struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns"`
}

type MySQLIndex struct {
	Name    string           `json:"name"`
	Unique  bool             `json:"unique"`
	Kind    MySQLIndexKind   `json:"kind"`
	Parts   []MySQLIndexPart `json:"parts"`
	Visible bool             `json:"visible"`
}

type MySQLIndexKind string

const (
	MySQLIndexBTree    MySQLIndexKind = "btree"
	MySQLIndexHash     MySQLIndexKind = "hash"
	MySQLIndexFullText MySQLIndexKind = "fulltext"
	MySQLIndexSpatial  MySQLIndexKind = "spatial"
	MySQLIndexOther    MySQLIndexKind = "other"
)

type MySQLIndexPart struct {
	Column       string `json:"column,omitempty"`
	Expression   string `json:"expression,omitempty"`
	PrefixLength int    `json:"prefixLength,omitempty"`
	Descending   bool   `json:"descending"`
}

type MySQLForeignKey struct {
	Name            string                  `json:"name,omitempty"`
	Columns         []MySQLForeignKeyColumn `json:"columns"`
	ReferencedTable MySQLTableReference     `json:"referencedTable"`
	OnUpdate        MySQLReferentialAction  `json:"onUpdate"`
	OnDelete        MySQLReferentialAction  `json:"onDelete"`
}

type MySQLForeignKeyColumn struct {
	Column           string `json:"column"`
	ReferencedColumn string `json:"referencedColumn"`
}

type MySQLTableReference struct {
	Database string `json:"database"`
	Table    string `json:"table"`
}

type MySQLReferentialAction string

const (
	MySQLNoAction MySQLReferentialAction = "no-action"
	MySQLRestrict MySQLReferentialAction = "restrict"
	MySQLCascade  MySQLReferentialAction = "cascade"
	MySQLSetNull  MySQLReferentialAction = "set-null"
)

type MySQLRelationship struct {
	FromTable   MySQLTableReference `json:"fromTable"`
	FromColumns []string            `json:"fromColumns"`
	ToTable     MySQLTableReference `json:"toTable"`
	ToColumns   []string            `json:"toColumns"`
	ForeignKey  string              `json:"foreignKey,omitempty"`
	Cardinality string              `json:"cardinality"`
	Optional    bool                `json:"optional"`
}

func (MySQLInspection) databaseInspection() {}

func convertMySQLRealm(realm *schema.Realm) *MySQLInspection {
	inspection := &MySQLInspection{
		Engine:        DatabaseEngineMySQL,
		Databases:     make([]MySQLDatabase, 0, len(realm.Schemas)),
		Relationships: make([]MySQLRelationship, 0),
	}
	for _, atlasDatabase := range realm.Schemas {
		if atlasDatabase == nil {
			continue
		}
		database := MySQLDatabase{
			Name:         atlasDatabase.Name,
			CharacterSet: databaseCharset(atlasDatabase.Attrs),
			Collation:    databaseCollation(atlasDatabase.Attrs),
			Tables:       make([]MySQLTable, 0, len(atlasDatabase.Tables)),
		}
		for _, atlasTable := range atlasDatabase.Tables {
			if atlasTable == nil {
				continue
			}
			database.Tables = append(database.Tables, convertMySQLTable(atlasTable))
			inspection.Relationships = append(
				inspection.Relationships,
				convertMySQLRelationships(realm, atlasTable)...,
			)
		}
		inspection.Databases = append(inspection.Databases, database)
	}
	return inspection
}

func convertMySQLTable(table *schema.Table) MySQLTable {
	converted := MySQLTable{
		Name:          table.Name,
		CharacterSet:  databaseCharset(table.Attrs),
		Collation:     databaseCollation(table.Attrs),
		Comment:       databaseComment(table.Attrs),
		Columns:       make([]MySQLColumn, 0, len(table.Columns)),
		Indexes:       make([]MySQLIndex, 0, len(table.Indexes)),
		ForeignKeys:   make([]MySQLForeignKey, 0, len(table.ForeignKeys)),
		StorageEngine: mysqlStorageEngine(table.Attrs),
	}
	for position, column := range table.Columns {
		if column != nil {
			converted.Columns = append(converted.Columns, convertMySQLColumn(column, position+1))
		}
	}
	if table.PrimaryKey != nil {
		converted.PrimaryKey = &MySQLPrimaryKey{
			Name:    table.PrimaryKey.Name,
			Columns: databaseIndexColumnNames(table.PrimaryKey),
		}
	}
	for _, index := range table.Indexes {
		if index != nil {
			converted.Indexes = append(converted.Indexes, convertMySQLIndex(index))
		}
	}
	for _, foreignKey := range table.ForeignKeys {
		if foreignKey != nil {
			converted.ForeignKeys = append(converted.ForeignKeys, convertMySQLForeignKey(foreignKey))
		}
	}
	return converted
}

func mysqlStorageEngine(attrs []schema.Attr) string {
	if engine, ok := databaseAttribute[*atlasmysql.Engine](attrs); ok {
		return engine.V
	}
	return ""
}

func convertMySQLColumn(column *schema.Column, position int) MySQLColumn {
	converted := MySQLColumn{
		Name:                column.Name,
		Position:            position,
		Type:                convertMySQLType(column.Type),
		DefaultExpression:   databaseExpression(column.Default),
		GeneratedExpression: databaseGeneratedExpression(column.Attrs),
		CharacterSet:        databaseCharset(column.Attrs),
		Collation:           databaseCollation(column.Attrs),
		Comment:             databaseComment(column.Attrs),
	}
	if column.Type != nil {
		converted.Nullable = column.Type.Null
	}
	_, converted.AutoIncrement = databaseAttribute[*atlasmysql.AutoIncrement](column.Attrs)
	return converted
}

func convertMySQLType(columnType *schema.ColumnType) MySQLType {
	if columnType == nil {
		return MySQLType{EnumValues: []string{}, SetValues: []string{}}
	}
	converted := MySQLType{
		NativeType: columnType.Raw,
		EnumValues: []string{},
		SetValues:  []string{},
	}
	if converted.NativeType == "" && columnType.Type != nil {
		if nativeType, err := atlasmysql.FormatType(columnType.Type); err == nil {
			converted.NativeType = nativeType
		}
	}
	populateMySQLType(&converted, columnType.Type)
	if converted.Name == "" {
		converted.Name = converted.NativeType
	}
	return converted
}

func populateMySQLType(converted *MySQLType, atlasType schema.Type) {
	switch atlasType := atlasType.(type) {
	case *schema.EnumType:
		converted.Name = atlasType.T
		converted.EnumValues = append([]string(nil), atlasType.Values...)
	case *atlasmysql.SetType:
		converted.Name = "set"
		converted.SetValues = append([]string(nil), atlasType.Values...)
	case *atlasmysql.BitType:
		converted.Name = atlasType.T
		converted.Length = atlasType.Size
	case *atlasmysql.NetworkType:
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
		converted.Unsigned = atlasType.Unsigned
	case *schema.DecimalType:
		converted.Name = atlasType.T
		converted.Precision = atlasType.Precision
		converted.Scale = atlasType.Scale
		converted.Unsigned = atlasType.Unsigned
	case *schema.FloatType:
		converted.Name = atlasType.T
		converted.Precision = atlasType.Precision
		converted.Unsigned = atlasType.Unsigned
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

func convertMySQLIndex(index *schema.Index) MySQLIndex {
	converted := MySQLIndex{
		Name:    index.Name,
		Unique:  index.Unique,
		Kind:    MySQLIndexBTree,
		Parts:   make([]MySQLIndexPart, 0, len(index.Parts)),
		Visible: true,
	}
	if indexType, ok := databaseAttribute[*atlasmysql.IndexType](index.Attrs); ok {
		converted.Kind = mysqlIndexKind(indexType.T)
	}
	for _, part := range index.Parts {
		if part == nil {
			continue
		}
		convertedPart := MySQLIndexPart{
			Expression: databaseIndexPartExpression(part),
			Descending: part.Desc,
		}
		if part.C != nil {
			convertedPart.Column = part.C.Name
		}
		if subPart, ok := databaseAttribute[*atlasmysql.SubPart](part.Attrs); ok {
			convertedPart.PrefixLength = subPart.Len
		}
		converted.Parts = append(converted.Parts, convertedPart)
	}
	return converted
}

func mysqlIndexKind(kind string) MySQLIndexKind {
	switch strings.ToUpper(kind) {
	case "", atlasmysql.IndexTypeBTree:
		return MySQLIndexBTree
	case atlasmysql.IndexTypeHash:
		return MySQLIndexHash
	case atlasmysql.IndexTypeFullText:
		return MySQLIndexFullText
	case atlasmysql.IndexTypeSpatial, "RTREE":
		return MySQLIndexSpatial
	default:
		return MySQLIndexOther
	}
}

func convertMySQLForeignKey(foreignKey *schema.ForeignKey) MySQLForeignKey {
	converted := MySQLForeignKey{
		Name:            foreignKey.Symbol,
		Columns:         make([]MySQLForeignKeyColumn, 0, len(foreignKey.Columns)),
		ReferencedTable: mysqlTableReference(foreignKey.RefTable),
		OnUpdate:        MySQLReferentialAction(databaseReferenceAction(foreignKey.OnUpdate)),
		OnDelete:        MySQLReferentialAction(databaseReferenceAction(foreignKey.OnDelete)),
	}
	for position, column := range foreignKey.Columns {
		if column == nil || position >= len(foreignKey.RefColumns) || foreignKey.RefColumns[position] == nil {
			continue
		}
		converted.Columns = append(converted.Columns, MySQLForeignKeyColumn{
			Column:           column.Name,
			ReferencedColumn: foreignKey.RefColumns[position].Name,
		})
	}
	return converted
}

func convertMySQLRelationships(realm *schema.Realm, table *schema.Table) []MySQLRelationship {
	relationships := make([]MySQLRelationship, 0, len(table.ForeignKeys))
	for _, foreignKey := range table.ForeignKeys {
		if foreignKey == nil || !databaseTableInRealm(realm, foreignKey.RefTable) {
			continue
		}
		relationships = append(relationships, MySQLRelationship{
			FromTable:   mysqlTableReference(table),
			FromColumns: databaseColumnNames(foreignKey.Columns),
			ToTable:     mysqlTableReference(foreignKey.RefTable),
			ToColumns:   databaseColumnNames(foreignKey.RefColumns),
			ForeignKey:  foreignKey.Symbol,
			Cardinality: databaseForeignKeyCardinality(table, foreignKey.Columns, func(*schema.Index) bool { return false }),
			Optional:    databaseForeignKeyOptional(foreignKey.Columns),
		})
	}
	return relationships
}

func mysqlTableReference(table *schema.Table) MySQLTableReference {
	reference := MySQLTableReference{}
	if table == nil {
		return reference
	}
	reference.Table = table.Name
	if table.Schema != nil {
		reference.Database = table.Schema.Name
	}
	return reference
}
