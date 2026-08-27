package ccode

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_ParseDatabaseURL(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		engine DatabaseEngine
	}{
		{name: "PostgreSQL", url: "postgres://localhost/application", engine: DatabaseEnginePostgreSQL},
		{name: "PostgreSQL alias", url: "postgresql://localhost/application", engine: DatabaseEnginePostgreSQL},
		{name: "MySQL", url: "mysql://localhost/application", engine: DatabaseEngineMySQL},
		{name: "MariaDB", url: "mariadb://localhost/application", engine: DatabaseEngineMariaDB},
		{name: "MariaDB short alias", url: "maria://localhost/application", engine: DatabaseEngineMariaDB},
		{name: "SQLite", url: "sqlite://database.db", engine: DatabaseEngineSQLite},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, engine, err := parseDatabaseURL(test.url)
			require.NoError(t, err)
			assert.Equal(t, test.engine, engine)
		})
	}
}

func TestDatabase_ParseDatabaseURLRejectsUnsupportedSchemeWithoutCredentials(t *testing.T) {
	_, _, err := parseDatabaseURL("oracle://admin:very-secret@localhost/application")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unsupported connection URL scheme "oracle"`)
	assert.NotContains(t, err.Error(), "very-secret")
}

func TestDatabase_InspectDatabaseValidatesArguments(t *testing.T) {
	inspection, err := InspectDatabase(nil, "sqlite://database.db")
	require.Error(t, err)
	assert.Nil(t, inspection)
	assert.Equal(t, "inspect database: context is nil", err.Error())

	inspection, err = InspectDatabase(context.Background(), "  ")
	require.Error(t, err)
	assert.Nil(t, inspection)
	assert.Equal(t, "inspect database: connection URL is blank", err.Error())
}

func TestDatabase_ConvertDatabaseRealmRejectsNilRealm(t *testing.T) {
	inspection, err := convertDatabaseRealm(DatabaseEngineSQLite, nil, nil)
	require.Error(t, err)
	assert.Nil(t, inspection)
	assert.Contains(t, err.Error(), "Atlas returned a nil realm")
}

func TestDatabase_DatabaseNameFromURL(t *testing.T) {
	parsedURL, _, err := parseDatabaseURL("postgres://localhost/customer%20portal?sslmode=disable")
	require.NoError(t, err)
	assert.Equal(t, "customer portal", databaseNameFromURL(parsedURL))
}

func TestDatabase_InspectSQLiteDatabase(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "application.db")
	db, err := sql.Open("sqlite3", databasePath)
	require.NoError(t, err)
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL
		);
		CREATE TABLE profiles (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL UNIQUE,
			display_name TEXT DEFAULT 'anonymous',
			CONSTRAINT profiles_user_fk
				FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		);
	`)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	inspection, err := InspectDatabase(context.Background(), "sqlite://"+databasePath)
	require.NoError(t, err)

	sqliteInspection, ok := inspection.(*SQLiteInspection)
	require.True(t, ok)
	assert.Equal(t, DatabaseEngineSQLite, sqliteInspection.Engine)
	require.Len(t, sqliteInspection.Databases, 1)
	assert.Equal(t, "main", sqliteInspection.Databases[0].Name)
	require.Len(t, sqliteInspection.Databases[0].Tables, 2)
	require.Len(t, sqliteInspection.Databases[0].Relationships, 1)
	assert.Equal(t, "one-to-one", sqliteInspection.Databases[0].Relationships[0].Cardinality)
	assert.False(t, sqliteInspection.Databases[0].Relationships[0].Optional)
}

func TestDatabaseTypeScript_InspectDatabaseReturnsGojaObject(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "application.db")
	createTypeScriptTestDatabase(t, databasePath)
	ctx := newRunnerDatabaseTestContext(t)

	value, err := ctx.InspectDatabase(
		"sqlite://"+databasePath,
		ctx.runtime.ToValue(map[string]any{"expectedEngine": "sqlite"}),
	)
	require.NoError(t, err)

	inspection := value.ToObject(ctx.runtime)
	assert.Equal(t, "sqlite", inspection.Get("engine").String())
	databases := inspection.Get("databases").ToObject(ctx.runtime)
	assert.Equal(t, int64(1), databases.Get("length").ToInteger())
	mainDatabase := databases.Get("0").ToObject(ctx.runtime)
	assert.Equal(t, "main", mainDatabase.Get("name").String())
	assert.Equal(t, int64(2), mainDatabase.Get("tables").ToObject(ctx.runtime).Get("length").ToInteger())
}

func TestDatabaseTypeScript_InspectDatabaseRejectsExpectedEngineMismatch(t *testing.T) {
	ctx := newRunnerDatabaseTestContext(t)

	value, err := ctx.InspectDatabase(
		"sqlite://database.db",
		ctx.runtime.ToValue(map[string]any{"expectedEngine": "mysql"}),
	)
	require.Error(t, err)
	assert.Nil(t, value)
	assert.Equal(t, "expected mysql database, but connection URL uses sqlite", err.Error())
}

func TestDatabaseTypeScript_InspectDatabaseValidatesOptions(t *testing.T) {
	ctx := newRunnerDatabaseTestContext(t)

	tests := []struct {
		name    string
		options []goja.Value
		message string
	}{
		{
			name:    "missing expected engine",
			options: []goja.Value{ctx.runtime.ToValue(map[string]any{})},
			message: "inspect database options must include expectedEngine",
		},
		{
			name:    "non-string expected engine",
			options: []goja.Value{ctx.runtime.ToValue(map[string]any{"expectedEngine": true})},
			message: "inspect database expectedEngine must be a string",
		},
		{
			name:    "unsupported expected engine",
			options: []goja.Value{ctx.runtime.ToValue(map[string]any{"expectedEngine": "oracle"})},
			message: `unsupported expected database engine "oracle"`,
		},
		{
			name:    "too many options",
			options: []goja.Value{ctx.runtime.NewObject(), ctx.runtime.NewObject()},
			message: "inspect database accepts at most one options argument",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := ctx.InspectDatabase("sqlite://database.db", test.options...)
			require.Error(t, err)
			assert.Nil(t, value)
			assert.Equal(t, test.message, err.Error())
		})
	}
}

func TestDatabaseTypeScript_RunnerInspectsDatabaseAndUsesTypeGuards(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestDatabaseTypeScript_RunnerInspectsDatabaseAndUsesTypeGuards")
	databasePath := filepath.Join(t.TempDir(), "application.db")
	createTypeScriptTestDatabase(t, databasePath)
	processFile := filepath.Join(projectDir, "database", "inspect.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	process := fmt.Sprintf(`import type { Context } from "@ccode/context";
import * as Database from "@ccode/database";

export default function main(ctx: Context) {
  const literalInspection = ctx.inspectDatabase(%q);
  const dynamicURL: string = %q;
  const dynamicInspection = ctx.inspectDatabase(dynamicURL, {
    expectedEngine: "sqlite",
  });

  if (!Database.SQLite.isInspection(dynamicInspection)) {
    throw new Error("expected SQLite inspection");
  }

  const guardResults = [
    Database.PostgreSQL.isInspection(
      { engine: "postgresql" } as Database.Inspection,
    ),
    Database.MySQL.isInspection({ engine: "mysql" } as Database.Inspection),
    Database.MariaDB.isInspection(
      { engine: "mariadb" } as Database.Inspection,
    ),
    Database.SQLite.isInspection({ engine: "sqlite" } as Database.Inspection),
  ];

  ctx.println(JSON.stringify({
    literalEngine: literalInspection.engine,
    database: dynamicInspection.databases[0].name,
    tableCount: dynamicInspection.databases[0].tables.length,
    guardResults,
  }));
}
`, "sqlite://"+databasePath, "sqlite://"+databasePath)
	require.NoError(t, os.WriteFile(processFile, []byte(process), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("database/inspect"))
	assert.JSONEq(t, `{
		"literalEngine": "sqlite",
		"database": "main",
		"tableCount": 2,
		"guardResults": [true, true, true, true]
	}`, output.String())
}

func newRunnerDatabaseTestContext(t *testing.T) *RunnerContext {
	t.Helper()

	runtime := goja.New()
	ctx := &RunnerContext{
		ccodeContext: NewContext(&Config{CCodePath: t.TempDir()}),
		runtime:      runtime,
	}
	require.NoError(t, ctx.initializeJSONParser())
	return ctx
}

func createTypeScriptTestDatabase(t *testing.T, databasePath string) {
	t.Helper()

	db, err := sql.Open("sqlite3", databasePath)
	require.NoError(t, err)
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL
		);
		CREATE TABLE profiles (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL UNIQUE,
			FOREIGN KEY (user_id) REFERENCES users (id)
		);
	`)
	require.NoError(t, err)
	require.NoError(t, db.Close())
}
