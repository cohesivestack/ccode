package ccode

import (
	"testing"

	"ariga.io/atlas/sql/schema"
	atlassqlite "ariga.io/atlas/sql/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_ConvertSQLiteRealm(t *testing.T) {
	main := schema.New("main")
	userID := sqliteTestColumn("id", "INTEGER", &schema.IntegerType{T: "integer"}, false)
	users := schema.NewTable("users").
		AddColumns(userID).
		SetPrimaryKey(schema.NewPrimaryKey(userID))

	profileID := sqliteTestColumn("id", "INTEGER", &schema.IntegerType{T: "integer"}, false)
	profileUserID := sqliteTestColumn("user_id", "INTEGER", &schema.IntegerType{T: "integer"}, true)
	displayName := sqliteTestColumn("display_name", "VARCHAR(120)", &schema.StringType{T: "varchar", Size: 120}, false).
		SetDefault(&schema.Literal{V: "'anonymous'"}).
		SetGeneratedExpr(&schema.GeneratedExpr{Expr: "upper(display_name)", Type: "VIRTUAL"})
	profiles := schema.NewTable("profiles").
		AddAttrs(&atlassqlite.Strict{}, &atlassqlite.WithoutRowID{}).
		AddColumns(profileID, profileUserID, displayName).
		SetPrimaryKey(schema.NewPrimaryKey(profileID))
	profiles.AddIndexes(
		schema.NewUniqueIndex("profiles_user_active_unique").
			AddColumns(profileUserID).
			AddAttrs(&atlassqlite.IndexPredicate{P: "user_id IS NOT NULL"}),
	)
	profiles.AddForeignKeys(
		schema.NewForeignKey("").
			AddColumns(profileUserID).
			SetRefTable(users).
			AddRefColumns(userID).
			SetOnDelete(schema.Cascade),
	)
	main.AddTables(users, profiles)

	inspection := convertSQLiteRealm(schema.NewRealm(main))

	assert.Equal(t, DatabaseEngineSQLite, inspection.Engine)
	require.Len(t, inspection.Databases, 1)
	require.Len(t, inspection.Databases[0].Tables, 2)
	convertedProfiles := inspection.Databases[0].Tables[1]
	assert.True(t, convertedProfiles.Strict)
	assert.True(t, convertedProfiles.WithoutRowID)
	assert.Equal(t, SQLiteAffinityInteger, convertedProfiles.Columns[1].Type.Affinity)
	assert.Equal(t, SQLiteAffinityText, convertedProfiles.Columns[2].Type.Affinity)
	assert.Equal(t, "'anonymous'", convertedProfiles.Columns[2].DefaultExpression)
	assert.Equal(t, "upper(display_name)", convertedProfiles.Columns[2].GeneratedExpression)
	require.Len(t, convertedProfiles.Indexes, 1)
	assert.Equal(t, "user_id IS NOT NULL", convertedProfiles.Indexes[0].Predicate)

	require.Len(t, inspection.Databases[0].Relationships, 1)
	relationship := inspection.Databases[0].Relationships[0]
	assert.Equal(t, "many-to-one", relationship.Cardinality)
	assert.True(t, relationship.Optional)
}

func TestDatabase_SQLiteTypeAffinity(t *testing.T) {
	tests := []struct {
		declaredType string
		affinity     SQLiteAffinity
	}{
		{declaredType: "BIGINT", affinity: SQLiteAffinityInteger},
		{declaredType: "VARCHAR(255)", affinity: SQLiteAffinityText},
		{declaredType: "BLOB", affinity: SQLiteAffinityBlob},
		{declaredType: "DOUBLE PRECISION", affinity: SQLiteAffinityReal},
		{declaredType: "DECIMAL(10,2)", affinity: SQLiteAffinityNumeric},
	}

	for _, test := range tests {
		t.Run(test.declaredType, func(t *testing.T) {
			assert.Equal(t, test.affinity, sqliteTypeAffinity(test.declaredType))
		})
	}
}

func sqliteTestColumn(name, rawType string, atlasType schema.Type, nullable bool) *schema.Column {
	column := schema.NewColumn(name).SetType(atlasType).SetNull(nullable)
	column.Type.Raw = rawType
	return column
}
