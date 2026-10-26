package ccode

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AddProcess creates a process under ccode_path without overwriting existing files.
func (ctx *Context) AddProcess(process string) error {
	if ctx == nil || ctx.config == nil || isStringBlank(ctx.config.CCodePath) {
		return fmt.Errorf("context config with ccode_path is required")
	}

	entryPointPath, err := processEntryPointPath(process)
	if err != nil {
		return err
	}
	sourcePath := filepath.Join(ctx.config.CCodePath, entryPointPath)
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		return fmt.Errorf("create process directory: %w", err)
	}

	file, err := os.OpenFile(sourcePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("create process %q: %w", process, err)
	}
	_, writeErr := file.WriteString("import type { Context } from \"@ccode/context\";\n\nexport default function run(ctx: Context) {\n\n}\n")
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write process %q: %w", process, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close process %q: %w", process, closeErr)
	}
	return nil
}

func processEntryPointPath(process string) (string, error) {
	if isStringBlank(process) {
		return "", fmt.Errorf("process name cannot be blank")
	}

	cleaned := filepath.Clean(filepath.FromSlash(process))
	if filepath.IsAbs(cleaned) || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid process path %q", process)
	}
	if strings.HasSuffix(cleaned, ".ts") && !strings.HasSuffix(cleaned, ProcessFileSuffix) {
		return "", fmt.Errorf("invalid process path %q: process files must end in %s or omit the suffix", process, ProcessFileSuffix)
	}
	if !strings.HasSuffix(cleaned, ProcessFileSuffix) {
		cleaned += ProcessFileSuffix
	}
	return cleaned, nil
}

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
