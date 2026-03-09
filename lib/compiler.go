package ccode

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/evanw/esbuild/pkg/api"
)

func (ctx *Context) compileTypescript(entryPointPath string) (api.BuildResult, error) {
	sourceHash, err := ctx.getSourceHash()
	if err != nil {
		return api.BuildResult{}, err
	}

	bundlePath := filepath.Join(ctx.config.HiddenPath, fmt.Sprintf("bundle.%s.js", sourceHash))
	sourceMapPath := bundlePath + ".map"

	if fileExists(bundlePath) && fileExists(sourceMapPath) {
		return buildResultFromPaths(bundlePath, sourceMapPath)
	}

	if err := os.MkdirAll(ctx.config.HiddenPath, 0755); err != nil {
		return api.BuildResult{}, fmt.Errorf("create hidden path: %w", err)
	}

	if err := removeExistingBundles(ctx.config.HiddenPath); err != nil {
		return api.BuildResult{}, err
	}

	result := api.Build(api.BuildOptions{
		AbsWorkingDir: ctx.config.Path,
		EntryPoints:   []string{entryPointPath},
		Bundle:        true,
		Platform:      api.PlatformBrowser, // or api.PlatformNode
		Format:        api.FormatESModule,  // or api.FormatCommonJS
		Sourcemap:     api.SourceMapLinked,
		Write:         true,
		Outfile:       bundlePath,
	})
	if len(result.Errors) > 0 {
		return result, nil
	}

	outputFiles, err := readOutputFiles(bundlePath, sourceMapPath)
	if err != nil {
		return result, err
	}
	result.OutputFiles = outputFiles
	return result, nil
}

func (ctx *Context) getSourceHash() (string, error) {
	var sourceFiles []string

	err := filepath.WalkDir(ctx.config.Path, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".ts" {
			return nil
		}
		sourceFiles = append(sourceFiles, path)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk TypeScript sources: %w", err)
	}
	if len(sourceFiles) == 0 {
		return "", fmt.Errorf("no TypeScript files found in %s", ctx.config.Path)
	}

	sort.Strings(sourceFiles)
	hasher := sha256.New()

	for _, sourceFile := range sourceFiles {
		relativePath, err := filepath.Rel(ctx.config.Path, sourceFile)
		if err != nil {
			return "", fmt.Errorf("build relative path for %s: %w", sourceFile, err)
		}

		contents, err := os.ReadFile(sourceFile)
		if err != nil {
			return "", fmt.Errorf("read TypeScript source %s: %w", sourceFile, err)
		}

		if _, err := hasher.Write([]byte(relativePath)); err != nil {
			return "", fmt.Errorf("hash source path %s: %w", sourceFile, err)
		}
		if _, err := hasher.Write([]byte{0}); err != nil {
			return "", fmt.Errorf("hash separator for %s: %w", sourceFile, err)
		}
		if _, err := hasher.Write(contents); err != nil {
			return "", fmt.Errorf("hash source contents %s: %w", sourceFile, err)
		}
		if _, err := hasher.Write([]byte{0}); err != nil {
			return "", fmt.Errorf("hash separator for %s: %w", sourceFile, err)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func buildResultFromPaths(paths ...string) (api.BuildResult, error) {
	outputFiles, err := readOutputFiles(paths...)
	if err != nil {
		return api.BuildResult{}, err
	}

	return api.BuildResult{OutputFiles: outputFiles}, nil
}

func readOutputFiles(paths ...string) ([]api.OutputFile, error) {
	outputFiles := make([]api.OutputFile, 0, len(paths))

	for _, path := range paths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read output file %s: %w", path, err)
		}

		outputFiles = append(outputFiles, api.OutputFile{
			Path:     path,
			Contents: contents,
		})
	}

	return outputFiles, nil
}

func removeExistingBundles(hiddenPath string) error {
	patterns := []string{
		filepath.Join(hiddenPath, "bundle.*.js"),
		filepath.Join(hiddenPath, "bundle.*.js.map"),
		filepath.Join(hiddenPath, "bundle.js"),
		filepath.Join(hiddenPath, "bundle.js.map"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("glob existing bundles with pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			if err := os.Remove(match); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove existing bundle %s: %w", match, err)
			}
		}
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
