package ccode

import (
	"testing"

	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_ConvertPostgreSQLRealm(t *testing.T) {
	public := schema.New("public")
	accountStatus := &schema.EnumType{
		T:      "account_status",
		Values: []string{"active", "disabled"},
		Schema: public,
	}
	id := schema.NewIntColumn("id", "bigint").
		AddAttrs(&postgres.Identity{Generation: "BY DEFAULT"})
	id.Type.Raw = "bigint"
	status := schema.NewColumn("status").
		SetType(accountStatus).
		SetDefault(&schema.RawExpr{X: "'active'::account_status"}).
		SetComment("Current account status")
	status.Type.Raw = "account_status"
	email := schema.NewStringColumn("email", "character varying", schema.StringSize(255))
	email.Type.Raw = "character varying(255)"
	users := schema.NewTable("users").
		SetComment("Application users").
		AddColumns(id, email, status).
		SetPrimaryKey(schema.NewPrimaryKey(id).SetName("users_pkey"))
	users.AddIndexes(
		schema.NewIndex("users_email_active_idx").
			AddColumns(email).
			AddAttrs(
				&postgres.IndexType{T: postgres.IndexTypeBTree},
				&postgres.IndexPredicate{P: "status = 'active'"},
				&postgres.IndexInclude{Columns: []*schema.Column{status}},
			),
	)
	public.AddObjects(accountStatus).AddTables(users)

	audit := schema.New("audit")
	eventID := schema.NewIntColumn("id", "bigint")
	userID := schema.NewNullIntColumn("user_id", "bigint")
	loginEvents := schema.NewTable("login_events").
		AddColumns(eventID, userID).
		SetPrimaryKey(schema.NewPrimaryKey(eventID).SetName("login_events_pkey"))
	loginEvents.AddForeignKeys(
		schema.NewForeignKey("login_events_user_fk").
			AddColumns(userID).
			SetRefTable(users).
			AddRefColumns(id).
			SetOnDelete(schema.SetNull),
	)
	audit.AddTables(loginEvents)

	inspection := convertPostgreSQLRealm(schema.NewRealm(public, audit), "application")

	assert.Equal(t, DatabaseEnginePostgreSQL, inspection.Engine)
	assert.Equal(t, "application", inspection.Database.Name)
	require.Len(t, inspection.Database.Schemas, 2)
	require.Len(t, inspection.Database.Schemas[0].EnumTypes, 1)
	assert.Equal(t, PostgreSQLEnumType{
		Name:   "account_status",
		Values: []string{"active", "disabled"},
	}, inspection.Database.Schemas[0].EnumTypes[0])

	convertedUsers := inspection.Database.Schemas[0].Tables[0]
	assert.Equal(t, "Application users", convertedUsers.Comment)
	require.NotNil(t, convertedUsers.PrimaryKey)
	assert.Equal(t, []string{"id"}, convertedUsers.PrimaryKey.Columns)
	require.Len(t, convertedUsers.Columns, 3)
	assert.Equal(t, "by-default", convertedUsers.Columns[0].Identity)
	assert.Equal(t, PostgreSQLType{
		Name:       "account_status",
		NativeType: "account_status",
		Schema:     "public",
	}, convertedUsers.Columns[2].Type)
	assert.Equal(t, "'active'::account_status", convertedUsers.Columns[2].DefaultExpression)
	require.Len(t, convertedUsers.Indexes, 1)
	assert.Equal(t, "btree", convertedUsers.Indexes[0].Method)
	assert.Equal(t, []string{"status"}, convertedUsers.Indexes[0].IncludedColumns)
	assert.Equal(t, "status = 'active'", convertedUsers.Indexes[0].Predicate)

	require.Len(t, inspection.Database.Relationships, 1)
	relationship := inspection.Database.Relationships[0]
	assert.Equal(t, PostgreSQLTableReference{Schema: "audit", Table: "login_events"}, relationship.FromTable)
	assert.Equal(t, PostgreSQLTableReference{Schema: "public", Table: "users"}, relationship.ToTable)
	assert.Equal(t, "many-to-one", relationship.Cardinality)
	assert.True(t, relationship.Optional)
}

func TestDatabase_ConvertPostgreSQLArrayType(t *testing.T) {
	atlasType := &postgres.ArrayType{
		T:    "text[][]",
		Type: &schema.StringType{T: "text"},
	}
	converted := convertPostgreSQLType(&schema.ColumnType{
		Raw:  "text[][]",
		Type: atlasType,
	})

	assert.Equal(t, "text", converted.Name)
	assert.Equal(t, "text[][]", converted.NativeType)
	assert.Equal(t, 2, converted.ArrayDimensions)
}
