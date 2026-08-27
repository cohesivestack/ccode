package templateassets

import (
	"embed"
	"io/fs"
	"sort"
	"strings"
)

//go:embed ccode.yaml.tpl
var ConfigTemplate string

//go:embed tsconfig.json
var TSConfigTemplate string

// SupportFS contains the TypeScript support API installed under .ccode/lib.
//
//go:embed context.ts openapi database
var SupportFS embed.FS

//go:embed ccode.gitignore
var HiddenGitIgnoreTemplate string

// SupportFilePaths returns every embedded TypeScript support file in stable
// path order.
func SupportFilePaths() ([]string, error) {
	paths := []string{}
	err := fs.WalkDir(SupportFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".ts") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(paths)
	return paths, nil
}
