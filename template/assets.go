package templateassets

import _ "embed"

//go:embed ccode.yaml.tpl
var ConfigTemplate string

//go:embed tsconfig.json
var TSConfigTemplate string

//go:embed types.ts
var TypesTemplate string
