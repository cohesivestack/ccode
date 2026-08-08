package ccode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlclient"
	"github.com/dop251/goja"
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

// InspectDatabase exposes database schema inspection to a TypeScript process.
func (ctx *RunnerContext) InspectDatabase(connectionURL string, optionValues ...goja.Value) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	expectedEngine, err := ctx.parseInspectDatabaseOptions(optionValues)
	if err != nil {
		return nil, err
	}
	if expectedEngine != "" {
		_, actualEngine, err := parseDatabaseURL(connectionURL)
		if err != nil {
			return nil, err
		}
		if actualEngine != expectedEngine {
			return nil, fmt.Errorf("expected %s database, but connection URL uses %s", expectedEngine, actualEngine)
		}
	}

	inspection, err := InspectDatabase(context.Background(), connectionURL)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(inspection)
	if err != nil {
		return nil, fmt.Errorf("serialize database inspection: %w", err)
	}
	return ctx.parseJSON(string(payload))
}

func (ctx *RunnerContext) parseInspectDatabaseOptions(optionValues []goja.Value) (DatabaseEngine, error) {
	if len(optionValues) == 0 {
		return "", nil
	}
	if len(optionValues) > 1 {
		return "", fmt.Errorf("inspect database accepts at most one options argument")
	}

	optionsValue := optionValues[0]
	if optionsValue == nil || goja.IsUndefined(optionsValue) || goja.IsNull(optionsValue) {
		return "", fmt.Errorf("inspect database options must be an object")
	}

	expectedEngineValue := optionsValue.ToObject(ctx.runtime).Get("expectedEngine")
	if expectedEngineValue == nil || goja.IsUndefined(expectedEngineValue) || goja.IsNull(expectedEngineValue) {
		return "", fmt.Errorf("inspect database options must include expectedEngine")
	}
	expectedEngine, ok := expectedEngineValue.Export().(string)
	if !ok {
		return "", fmt.Errorf("inspect database expectedEngine must be a string")
	}

	switch DatabaseEngine(expectedEngine) {
	case DatabaseEnginePostgreSQL, DatabaseEngineMySQL, DatabaseEngineMariaDB, DatabaseEngineSQLite:
		return DatabaseEngine(expectedEngine), nil
	default:
		return "", fmt.Errorf("unsupported expected database engine %q", expectedEngine)
	}
}
