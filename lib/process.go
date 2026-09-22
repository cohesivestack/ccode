package ccode

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// ListProcesses returns runnable process identifiers relative to ccode_path.
func (ctx *Context) ListProcesses() ([]string, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
	}

	rootPath := filepath.Clean(ctx.config.CCodePath)
	hiddenPath := filepath.Clean(ctx.config.HiddenPath)
	processes := []string{}

	err := filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			if path == hiddenPath || (path != rootPath && strings.HasPrefix(entry.Name(), ".")) {
				return fs.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ProcessFileSuffix) {
			return nil
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return fmt.Errorf("build process path for %s: %w", path, err)
		}

		processName := strings.TrimSuffix(filepath.ToSlash(relativePath), ProcessFileSuffix)
		if !isStringBlank(processName) {
			processes = append(processes, processName)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list processes in %s: %w", rootPath, err)
	}

	sort.Strings(processes)
	return processes, nil
}
