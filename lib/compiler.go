package ccode

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func (ctx *Context) compileTypescript(entryPointPath string) (api.BuildResult, error) {
	sourceHash, err := ctx.getSourceHash()
	if err != nil {
		return api.BuildResult{}, err
	}

	bundlePath := filepath.Join(ctx.config.HiddenPath, fmt.Sprintf("bundle.%s.js", sourceHash))
	sourceMapPath := bundlePath + ".map"

	return ctx.compileTypescriptBundle(
		entryPointPath,
		bundlePath,
		sourceMapPath,
		ctx.config.HiddenPath,
		[]string{
			filepath.Join(ctx.config.HiddenPath, "bundle.*.js"),
			filepath.Join(ctx.config.HiddenPath, "bundle.*.js.map"),
			filepath.Join(ctx.config.HiddenPath, "bundle.js"),
			filepath.Join(ctx.config.HiddenPath, "bundle.js.map"),
		},
		api.BuildOptions{
			AbsWorkingDir: ctx.config.Path,
			EntryPoints:   []string{entryPointPath},
			Bundle:        true,
			Platform:      api.PlatformBrowser,
			Format:        api.FormatESModule,
			Sourcemap:     api.SourceMapLinked,
			Write:         true,
			Outfile:       bundlePath,
		},
	)
}

func (ctx *Context) compileTypescriptForRunner(entryPointPath string) (api.BuildResult, error) {
	sourceHash, err := ctx.getSourceHash()
	if err != nil {
		return api.BuildResult{}, err
	}

	entryPointHash := hashString(entryPointPath)
	buildPath := filepath.Join(ctx.config.HiddenPath, "build")
	bundlePrefix := fmt.Sprintf("process.%s", entryPointHash)
	bundlePath := filepath.Join(buildPath, fmt.Sprintf("%s.%s.js", bundlePrefix, sourceHash))
	sourceMapPath := bundlePath + ".map"

	return ctx.compileTypescriptBundle(
		entryPointPath,
		bundlePath,
		sourceMapPath,
		buildPath,
		[]string{
			filepath.Join(buildPath, fmt.Sprintf("%s.*.js", bundlePrefix)),
			filepath.Join(buildPath, fmt.Sprintf("%s.*.js.map", bundlePrefix)),
		},
		api.BuildOptions{
			AbsWorkingDir: ctx.config.Path,
			EntryPoints:   []string{entryPointPath},
			Bundle:        true,
			Platform:      api.PlatformNeutral,
			Format:        api.FormatCommonJS,
			Sourcemap:     api.SourceMapLinked,
			Write:         true,
			Outfile:       bundlePath,
		},
	)
}

func (ctx *Context) compileTypescriptBundle(
	entryPointPath string,
	bundlePath string,
	sourceMapPath string,
	outputDir string,
	cleanupPatterns []string,
	options api.BuildOptions,
) (api.BuildResult, error) {
	if fileExists(bundlePath) && fileExists(sourceMapPath) {
		return buildResultFromPaths(bundlePath, sourceMapPath)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return api.BuildResult{}, fmt.Errorf("create output path %s: %w", outputDir, err)
	}

	if err := removeMatchingFiles(cleanupPatterns...); err != nil {
		return api.BuildResult{}, err
	}

	result := api.Build(options)
	if len(result.Errors) > 0 {
		return result, fmt.Errorf("compile %s: %s", entryPointPath, formatBuildMessages(result.Errors))
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

func removeMatchingFiles(patterns ...string) error {
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("glob files with pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			if err := os.Remove(match); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove file %s: %w", match, err)
			}
		}
	}

	return nil
}

func formatBuildMessages(messages []api.Message) string {
	if len(messages) == 0 {
		return "unknown build error"
	}

	formatted := make([]string, 0, len(messages))
	for _, message := range messages {
		text := message.Text
		if message.Location == nil {
			formatted = append(formatted, text)
			continue
		}

		formatted = append(formatted, fmt.Sprintf("%s:%d:%d: %s", message.Location.File, message.Location.Line, message.Location.Column, text))
	}

	return strings.Join(formatted, "; ")
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
