package ccode

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlclient"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseEngine identifies a database engine supported by schema inspection.
type DatabaseEngine string

const (
	DatabaseEnginePostgreSQL DatabaseEngine = "postgresql"
	DatabaseEngineMySQL      DatabaseEngine = "mysql"
	DatabaseEngineMariaDB    DatabaseEngine = "mariadb"
	DatabaseEngineSQLite     DatabaseEngine = "sqlite"
)

// DatabaseInspection is implemented by every engine-specific inspection model.
type DatabaseInspection interface {
	databaseInspection()
}

// InspectDatabase reads the schema reachable through connectionURL and converts
// it to the Cohesive Code model for the URL's database engine.
func InspectDatabase(ctx context.Context, connectionURL string) (DatabaseInspection, error) {
	if ctx == nil {
		return nil, fmt.Errorf("inspect database: context is nil")
	}
	if isStringBlank(connectionURL) {
		return nil, fmt.Errorf("inspect database: connection URL is blank")
	}

	parsedURL, engine, err := parseDatabaseURL(connectionURL)
	if err != nil {
		return nil, err
	}

	client, err := sqlclient.Open(ctx, connectionURL)
	if err != nil {
		return nil, fmt.Errorf("inspect %s database: open connection: %w", engine, err)
	}

	realm, err := client.InspectRealm(ctx, nil)
	closeErr := client.Close()
	if err != nil {
		return nil, fmt.Errorf("inspect %s database schema: %w", engine, err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("inspect %s database: close connection: %w", engine, closeErr)
	}

	return convertDatabaseRealm(engine, parsedURL, realm)
}

func parseDatabaseURL(connectionURL string) (*url.URL, DatabaseEngine, error) {
	parsedURL, err := sqlclient.ParseURL(connectionURL)
	if err != nil {
		return nil, "", fmt.Errorf("inspect database: parse connection URL: %w", err)
	}

	var engine DatabaseEngine
	switch strings.ToLower(parsedURL.Scheme) {
	case "postgres", "postgresql":
		engine = DatabaseEnginePostgreSQL
	case "mysql":
		engine = DatabaseEngineMySQL
	case "maria", "mariadb":
		engine = DatabaseEngineMariaDB
	case "sqlite":
		engine = DatabaseEngineSQLite
	default:
		return nil, "", fmt.Errorf("inspect database: unsupported connection URL scheme %q", parsedURL.Scheme)
	}
	return parsedURL, engine, nil
}

func convertDatabaseRealm(engine DatabaseEngine, connectionURL *url.URL, realm *schema.Realm) (DatabaseInspection, error) {
	if realm == nil {
		return nil, fmt.Errorf("inspect %s database: Atlas returned a nil realm", engine)
	}

	switch engine {
	case DatabaseEnginePostgreSQL:
		return convertPostgreSQLRealm(realm, databaseNameFromURL(connectionURL)), nil
	case DatabaseEngineMySQL:
		return convertMySQLRealm(realm), nil
	case DatabaseEngineMariaDB:
		return convertMariaDBRealm(realm), nil
	case DatabaseEngineSQLite:
		return convertSQLiteRealm(realm), nil
	default:
		return nil, fmt.Errorf("inspect database: unsupported engine %q", engine)
	}
}

func databaseNameFromURL(connectionURL *url.URL) string {
	if connectionURL == nil {
		return ""
	}
	name := strings.Trim(connectionURL.EscapedPath(), "/")
	if decoded, err := url.PathUnescape(name); err == nil {
		return decoded
	}
	return name
}
