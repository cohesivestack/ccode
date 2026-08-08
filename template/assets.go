package templateassets

import _ "embed"

//go:embed ccode.yaml.tpl
var ConfigTemplate string

//go:embed tsconfig.json
var TSConfigTemplate string

//go:embed context.ts
var ContextTemplate string

//go:embed openapi.ts
var OpenAPITemplate string

//go:embed database.ts
var DatabaseTemplate string

//go:embed database-postgresql.ts
var DatabasePostgreSQLTemplate string

//go:embed database-mysql.ts
var DatabaseMySQLTemplate string

//go:embed database-mariadb.ts
var DatabaseMariaDBTemplate string

//go:embed database-sqlite.ts
var DatabaseSQLiteTemplate string

//go:embed ccode.gitignore
var HiddenGitIgnoreTemplate string
