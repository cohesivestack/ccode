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

//go:embed ccode.gitignore
var HiddenGitIgnoreTemplate string
