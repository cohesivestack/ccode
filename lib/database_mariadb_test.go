package ccode

import (
	"testing"

	atlasmysql "ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_ConvertMariaDBRealm(t *testing.T) {
	database := schema.New("application").
		SetCharset("utf8mb4").
		SetCollation("utf8mb4_unicode_ci")
	id := schema.NewIntColumn("id", "bigint").AddAttrs(&atlasmysql.AutoIncrement{})
	role := schema.NewEnumColumn(
		"role",
		schema.EnumName("enum"),
		schema.EnumValues("admin", "member"),
	)
	role.Type.Raw = "enum('admin','member')"
	users := schema.NewTable("users").
		AddAttrs(&atlasmysql.Engine{V: "InnoDB"}).
		AddColumns(id, role).
		SetPrimaryKey(schema.NewPrimaryKey(id).SetName("PRIMARY"))
	users.AddIndexes(
		schema.NewIndex("users_role_fulltext").
			AddColumns(role).
			AddAttrs(&atlasmysql.IndexType{T: atlasmysql.IndexTypeFullText}),
	)
	database.AddTables(users)

	inspection := convertMariaDBRealm(schema.NewRealm(database))

	assert.Equal(t, DatabaseEngineMariaDB, inspection.Engine)
	require.Len(t, inspection.Databases, 1)
	assert.Equal(t, "utf8mb4_unicode_ci", inspection.Databases[0].Collation)
	require.Len(t, inspection.Databases[0].Tables, 1)
	convertedUsers := inspection.Databases[0].Tables[0]
	assert.Equal(t, "InnoDB", convertedUsers.StorageEngine)
	assert.True(t, convertedUsers.Columns[0].AutoIncrement)
	assert.Equal(t, []string{"admin", "member"}, convertedUsers.Columns[1].Type.EnumValues)
	require.Len(t, convertedUsers.Indexes, 1)
	assert.Equal(t, MariaDBIndexFullText, convertedUsers.Indexes[0].Kind)
	assert.False(t, convertedUsers.Indexes[0].Ignored)
}

func TestDatabase_MariaDBModelIsIndependentFromMySQL(t *testing.T) {
	var inspection DatabaseInspection = &MariaDBInspection{}
	_, isMariaDB := inspection.(*MariaDBInspection)
	_, isMySQL := inspection.(*MySQLInspection)

	assert.True(t, isMariaDB)
	assert.False(t, isMySQL)
}
