package ccode

import (
	"strings"

	atlasmysql "ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/schema"
)

type MariaDBInspection struct {
	Engine        DatabaseEngine        `json:"engine"`
	Databases     []MariaDBDatabase     `json:"databases"`
	Relationships []MariaDBRelationship `json:"relationships"`
}

type MariaDBDatabase struct {
	Name         string         `json:"name"`
	CharacterSet string         `json:"characterSet,omitempty"`
	Collation    string         `json:"collation,omitempty"`
	Tables       []MariaDBTable `json:"tables"`
}

type MariaDBTable struct {
	Name          string              `json:"name"`
	StorageEngine string              `json:"storageEngine"`
	CharacterSet  string              `json:"characterSet,omitempty"`
	Collation     string              `json:"collation,omitempty"`
	Comment       string              `json:"comment,omitempty"`
	Columns       []MariaDBColumn     `json:"columns"`
	PrimaryKey    *MariaDBPrimaryKey  `json:"primaryKey,omitempty"`
	Indexes       []MariaDBIndex      `json:"indexes"`
	ForeignKeys   []MariaDBForeignKey `json:"foreignKeys"`
}

type MariaDBColumn struct {
	Name                string      `json:"name"`
	Position            int         `json:"position"`
	Type                MariaDBType `json:"type"`
	Nullable            bool        `json:"nullable"`
	DefaultExpression   string      `json:"defaultExpression,omitempty"`
	GeneratedExpression string      `json:"generatedExpression,omitempty"`
	AutoIncrement       bool        `json:"autoIncrement"`
	CharacterSet        string      `json:"characterSet,omitempty"`
	Collation           string      `json:"collation,omitempty"`
	Comment             string      `json:"comment,omitempty"`
}

type MariaDBType struct {
	Name       string   `json:"name"`
	NativeType string   `json:"nativeType"`
	Length     int      `json:"length,omitempty"`
	Precision  int      `json:"precision,omitempty"`
	Scale      int      `json:"scale,omitempty"`
	Unsigned   bool     `json:"unsigned"`
	EnumValues []string `json:"enumValues,omitempty"`
	SetValues  []string `json:"setValues,omitempty"`
}

type MariaDBPrimaryKey struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns"`
}

type MariaDBIndex struct {
	Name    string             `json:"name"`
	Unique  bool               `json:"unique"`
	Kind    MariaDBIndexKind   `json:"kind"`
	Parts   []MariaDBIndexPart `json:"parts"`
	Ignored bool               `json:"ignored"`
}

type MariaDBIndexKind string

const (
	MariaDBIndexBTree    MariaDBIndexKind = "btree"
	MariaDBIndexHash     MariaDBIndexKind = "hash"
	MariaDBIndexFullText MariaDBIndexKind = "fulltext"
	MariaDBIndexSpatial  MariaDBIndexKind = "spatial"
	MariaDBIndexOther    MariaDBIndexKind = "other"
)

type MariaDBIndexPart struct {
	Column       string `json:"column,omitempty"`
	Expression   string `json:"expression,omitempty"`
	PrefixLength int    `json:"prefixLength,omitempty"`
	Descending   bool   `json:"descending"`
}

type MariaDBForeignKey struct {
	Name            string                    `json:"name,omitempty"`
	Columns         []MariaDBForeignKeyColumn `json:"columns"`
	ReferencedTable MariaDBTableReference     `json:"referencedTable"`
	OnUpdate        MariaDBReferentialAction  `json:"onUpdate"`
	OnDelete        MariaDBReferentialAction  `json:"onDelete"`
}

type MariaDBForeignKeyColumn struct {
	Column           string `json:"column"`
	ReferencedColumn string `json:"referencedColumn"`
}

type MariaDBTableReference struct {
	Database string `json:"database"`
	Table    string `json:"table"`
}

type MariaDBReferentialAction string

const (
	MariaDBNoAction MariaDBReferentialAction = "no-action"
	MariaDBRestrict MariaDBReferentialAction = "restrict"
	MariaDBCascade  MariaDBReferentialAction = "cascade"
	MariaDBSetNull  MariaDBReferentialAction = "set-null"
)

type MariaDBRelationship struct {
	FromTable   MariaDBTableReference `json:"fromTable"`
	FromColumns []string              `json:"fromColumns"`
	ToTable     MariaDBTableReference `json:"toTable"`
	ToColumns   []string              `json:"toColumns"`
	ForeignKey  string                `json:"foreignKey,omitempty"`
	Cardinality string                `json:"cardinality"`
	Optional    bool                  `json:"optional"`
}

func (MariaDBInspection) databaseInspection() {}

func convertMariaDBRealm(realm *schema.Realm) *MariaDBInspection {
	inspection := &MariaDBInspection{
		Engine:        DatabaseEngineMariaDB,
		Databases:     make([]MariaDBDatabase, 0, len(realm.Schemas)),
		Relationships: make([]MariaDBRelationship, 0),
	}
	for _, atlasDatabase := range realm.Schemas {
		if atlasDatabase == nil {
			continue
		}
		database := MariaDBDatabase{
			Name:         atlasDatabase.Name,
			CharacterSet: databaseCharset(atlasDatabase.Attrs),
			Collation:    databaseCollation(atlasDatabase.Attrs),
			Tables:       make([]MariaDBTable, 0, len(atlasDatabase.Tables)),
		}
		for _, atlasTable := range atlasDatabase.Tables {
			if atlasTable == nil {
				continue
			}
			database.Tables = append(database.Tables, convertMariaDBTable(atlasTable))
			inspection.Relationships = append(
				inspection.Relationships,
				convertMariaDBRelationships(realm, atlasTable)...,
			)
		}
		inspection.Databases = append(inspection.Databases, database)
	}
	return inspection
}

func convertMariaDBTable(table *schema.Table) MariaDBTable {
	converted := MariaDBTable{
		Name:          table.Name,
		StorageEngine: mariaDBStorageEngine(table.Attrs),
		CharacterSet:  databaseCharset(table.Attrs),
		Collation:     databaseCollation(table.Attrs),
		Comment:       databaseComment(table.Attrs),
		Columns:       make([]MariaDBColumn, 0, len(table.Columns)),
		Indexes:       make([]MariaDBIndex, 0, len(table.Indexes)),
		ForeignKeys:   make([]MariaDBForeignKey, 0, len(table.ForeignKeys)),
	}
	for position, column := range table.Columns {
		if column != nil {
			converted.Columns = append(converted.Columns, convertMariaDBColumn(column, position+1))
		}
	}
	if table.PrimaryKey != nil {
		converted.PrimaryKey = &MariaDBPrimaryKey{
			Name:    table.PrimaryKey.Name,
			Columns: databaseIndexColumnNames(table.PrimaryKey),
		}
	}
	for _, index := range table.Indexes {
		if index != nil {
			converted.Indexes = append(converted.Indexes, convertMariaDBIndex(index))
		}
	}
	for _, foreignKey := range table.ForeignKeys {
		if foreignKey != nil {
			converted.ForeignKeys = append(converted.ForeignKeys, convertMariaDBForeignKey(foreignKey))
		}
	}
	return converted
}

func mariaDBStorageEngine(attrs []schema.Attr) string {
	if engine, ok := databaseAttribute[*atlasmysql.Engine](attrs); ok {
		return engine.V
	}
	return ""
}

func convertMariaDBColumn(column *schema.Column, position int) MariaDBColumn {
	converted := MariaDBColumn{
		Name:                column.Name,
		Position:            position,
		Type:                convertMariaDBType(column.Type),
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

func convertMariaDBType(columnType *schema.ColumnType) MariaDBType {
	if columnType == nil {
		return MariaDBType{EnumValues: []string{}, SetValues: []string{}}
	}
	converted := MariaDBType{
		NativeType: columnType.Raw,
		EnumValues: []string{},
		SetValues:  []string{},
	}
	if converted.NativeType == "" && columnType.Type != nil {
		if nativeType, err := atlasmysql.FormatType(columnType.Type); err == nil {
			converted.NativeType = nativeType
		}
	}
	populateMariaDBType(&converted, columnType.Type)
	if converted.Name == "" {
		converted.Name = converted.NativeType
	}
	return converted
}

func populateMariaDBType(converted *MariaDBType, atlasType schema.Type) {
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

func convertMariaDBIndex(index *schema.Index) MariaDBIndex {
	converted := MariaDBIndex{
		Name:   index.Name,
		Unique: index.Unique,
		Kind:   MariaDBIndexBTree,
		Parts:  make([]MariaDBIndexPart, 0, len(index.Parts)),
	}
	if indexType, ok := databaseAttribute[*atlasmysql.IndexType](index.Attrs); ok {
		converted.Kind = mariaDBIndexKind(indexType.T)
	}
	for _, part := range index.Parts {
		if part == nil {
			continue
		}
		convertedPart := MariaDBIndexPart{
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

func mariaDBIndexKind(kind string) MariaDBIndexKind {
	switch strings.ToUpper(kind) {
	case "", atlasmysql.IndexTypeBTree:
		return MariaDBIndexBTree
	case atlasmysql.IndexTypeHash:
		return MariaDBIndexHash
	case atlasmysql.IndexTypeFullText:
		return MariaDBIndexFullText
	case atlasmysql.IndexTypeSpatial, "RTREE":
		return MariaDBIndexSpatial
	default:
		return MariaDBIndexOther
	}
}

func convertMariaDBForeignKey(foreignKey *schema.ForeignKey) MariaDBForeignKey {
	converted := MariaDBForeignKey{
		Name:            foreignKey.Symbol,
		Columns:         make([]MariaDBForeignKeyColumn, 0, len(foreignKey.Columns)),
		ReferencedTable: mariaDBTableReference(foreignKey.RefTable),
		OnUpdate:        MariaDBReferentialAction(databaseReferenceAction(foreignKey.OnUpdate)),
		OnDelete:        MariaDBReferentialAction(databaseReferenceAction(foreignKey.OnDelete)),
	}
	for position, column := range foreignKey.Columns {
		if column == nil || position >= len(foreignKey.RefColumns) || foreignKey.RefColumns[position] == nil {
			continue
		}
		converted.Columns = append(converted.Columns, MariaDBForeignKeyColumn{
			Column:           column.Name,
			ReferencedColumn: foreignKey.RefColumns[position].Name,
		})
	}
	return converted
}

func convertMariaDBRelationships(realm *schema.Realm, table *schema.Table) []MariaDBRelationship {
	relationships := make([]MariaDBRelationship, 0, len(table.ForeignKeys))
	for _, foreignKey := range table.ForeignKeys {
		if foreignKey == nil || !databaseTableInRealm(realm, foreignKey.RefTable) {
			continue
		}
		relationships = append(relationships, MariaDBRelationship{
			FromTable:   mariaDBTableReference(table),
			FromColumns: databaseColumnNames(foreignKey.Columns),
			ToTable:     mariaDBTableReference(foreignKey.RefTable),
			ToColumns:   databaseColumnNames(foreignKey.RefColumns),
			ForeignKey:  foreignKey.Symbol,
			Cardinality: databaseForeignKeyCardinality(table, foreignKey.Columns, func(*schema.Index) bool { return false }),
			Optional:    databaseForeignKeyOptional(foreignKey.Columns),
		})
	}
	return relationships
}

func mariaDBTableReference(table *schema.Table) MariaDBTableReference {
	reference := MariaDBTableReference{}
	if table == nil {
		return reference
	}
	reference.Table = table.Name
	if table.Schema != nil {
		reference.Database = table.Schema.Name
	}
	return reference
}
