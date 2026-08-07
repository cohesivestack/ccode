package ccode

import (
	"testing"

	atlasmysql "ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_ConvertMySQLRealm(t *testing.T) {
	application := schema.New("application").
		SetCharset("utf8mb4").
		SetCollation("utf8mb4_0900_ai_ci")
	userID := schema.NewUintColumn("id", "bigint").AddAttrs(&atlasmysql.AutoIncrement{})
	userID.Type.Raw = "bigint unsigned"
	status := schema.NewEnumColumn(
		"status",
		schema.EnumName("enum"),
		schema.EnumValues("active", "disabled"),
	)
	status.Type.Raw = "enum('active','disabled')"
	users := schema.NewTable("users").
		AddAttrs(&atlasmysql.Engine{V: "InnoDB"}).
		SetCharset("utf8mb4").
		SetComment("Application users").
		AddColumns(userID, status).
		SetPrimaryKey(schema.NewPrimaryKey(userID).SetName("PRIMARY"))
	application.AddTables(users)

	authentication := schema.New("authentication")
	profileID := schema.NewUintColumn("id", "bigint")
	profileUserID := schema.NewUintColumn("user_id", "bigint")
	profileUserID.Type.Raw = "bigint unsigned"
	profiles := schema.NewTable("profiles").
		AddAttrs(&atlasmysql.Engine{V: "InnoDB"}).
		AddColumns(profileID, profileUserID).
		SetPrimaryKey(schema.NewPrimaryKey(profileID).SetName("PRIMARY"))
	profiles.AddIndexes(
		schema.NewUniqueIndex("profiles_user_unique").AddColumns(profileUserID),
		schema.NewIndex("profiles_user_prefix").AddParts(
			schema.NewColumnPart(profileUserID).
				SetDesc(true).
				AddAttrs(&atlasmysql.SubPart{Len: 8}),
		).AddAttrs(&atlasmysql.IndexType{T: atlasmysql.IndexTypeHash}),
	)
	profiles.AddForeignKeys(
		schema.NewForeignKey("profiles_user_fk").
			AddColumns(profileUserID).
			SetRefTable(users).
			AddRefColumns(userID).
			SetOnUpdate(schema.Cascade).
			SetOnDelete(schema.Cascade),
	)
	authentication.AddTables(profiles)

	inspection := convertMySQLRealm(schema.NewRealm(application, authentication))

	assert.Equal(t, DatabaseEngineMySQL, inspection.Engine)
	require.Len(t, inspection.Databases, 2)
	assert.Equal(t, "utf8mb4", inspection.Databases[0].CharacterSet)
	convertedUsers := inspection.Databases[0].Tables[0]
	assert.Equal(t, "InnoDB", convertedUsers.StorageEngine)
	assert.Equal(t, "Application users", convertedUsers.Comment)
	assert.True(t, convertedUsers.Columns[0].AutoIncrement)
	assert.True(t, convertedUsers.Columns[0].Type.Unsigned)
	assert.Equal(t, []string{"active", "disabled"}, convertedUsers.Columns[1].Type.EnumValues)

	convertedProfiles := inspection.Databases[1].Tables[0]
	require.Len(t, convertedProfiles.Indexes, 2)
	assert.Equal(t, MySQLIndexHash, convertedProfiles.Indexes[1].Kind)
	assert.Equal(t, 8, convertedProfiles.Indexes[1].Parts[0].PrefixLength)
	assert.True(t, convertedProfiles.Indexes[1].Parts[0].Descending)

	require.Len(t, inspection.Relationships, 1)
	relationship := inspection.Relationships[0]
	assert.Equal(t, MySQLTableReference{Database: "authentication", Table: "profiles"}, relationship.FromTable)
	assert.Equal(t, MySQLTableReference{Database: "application", Table: "users"}, relationship.ToTable)
	assert.Equal(t, "one-to-one", relationship.Cardinality)
	assert.False(t, relationship.Optional)
}

func TestDatabase_ConvertMySQLSetType(t *testing.T) {
	converted := convertMySQLType(&schema.ColumnType{
		Raw:  "set('read','write')",
		Type: &atlasmysql.SetType{Values: []string{"read", "write"}},
	})

	assert.Equal(t, "set", converted.Name)
	assert.Equal(t, []string{"read", "write"}, converted.SetValues)
}
