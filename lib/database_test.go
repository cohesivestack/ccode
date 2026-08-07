package ccode

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

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
