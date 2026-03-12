package ccode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCompiler_CompileTypescript(t *testing.T) {
	ctx, config, mainFile, helperFile := setupLoaderTestProject(t, "TestCompiler_CompileTypescript")

	require.NoError(t, os.WriteFile(helperFile, []byte(`export const message = "compiled helper";`), 0644))
	require.NoError(t, os.WriteFile(mainFile, []byte("import type { Context } from \"@ccode/context\";\nimport { message } from \"./helper\";\n\nexport default function main(ctx: Context) {\n\tctx.println(message);\n}\n"), 0644))

	result, err := ctx.compileTypescript("main.ts")
	require.NoError(t, err)
	require.Empty(t, result.Errors, "esbuild returned errors: %+v", result.Errors)

	sourceHash, err := ctx.getSourceHash()
	require.NoError(t, err)

	bundlePath := filepath.Join(config.HiddenPath, fmt.Sprintf("bundle.%s.js", sourceHash))
	sourceMapPath := bundlePath + ".map"

	bundleContent, err := os.ReadFile(bundlePath)
	require.NoError(t, err)
	require.FileExists(t, bundlePath)
	require.FileExists(t, sourceMapPath)
	require.Contains(t, string(bundleContent), "compiled helper")
	require.Contains(t, string(bundleContent), "function main(ctx)")
	require.Contains(t, string(bundleContent), "ctx.println(message)")
	require.Len(t, result.OutputFiles, 2)
	require.Equal(t, bundlePath, result.OutputFiles[0].Path)
	require.Equal(t, sourceMapPath, result.OutputFiles[1].Path)

	sourceMapContent, err := os.ReadFile(sourceMapPath)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(sourceMapContent), "main.ts"))
	require.True(t, strings.Contains(string(sourceMapContent), "helper.ts"))

	bundleInfoBefore, err := os.Stat(bundlePath)
	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)

	cachedResult, err := ctx.compileTypescript("main.ts")
	require.NoError(t, err)
	require.Empty(t, cachedResult.Errors)
	require.Len(t, cachedResult.OutputFiles, 2)

	bundleInfoAfter, err := os.Stat(bundlePath)
	require.NoError(t, err)
	require.True(t, bundleInfoBefore.ModTime().Equal(bundleInfoAfter.ModTime()))
}

func TestCompiler_CompileTypescript_RebuildsForChangedSources(t *testing.T) {
	ctx, config, mainFile, helperFile := setupLoaderTestProject(t, "TestCompiler_CompileTypescript_RebuildsForChangedSources")

	require.NoError(t, os.WriteFile(helperFile, []byte(`export const message = "first build";`), 0644))
	require.NoError(t, os.WriteFile(mainFile, []byte("import type { Context } from \"@ccode/context\";\nimport { message } from \"./helper\";\n\nexport default function main(ctx: Context) {\n\tctx.println(message);\n}\n"), 0644))

	_, err := ctx.compileTypescript("main.ts")
	require.NoError(t, err)

	firstHash, err := ctx.getSourceHash()
	require.NoError(t, err)

	firstBundlePath := filepath.Join(config.HiddenPath, fmt.Sprintf("bundle.%s.js", firstHash))
	firstSourceMapPath := firstBundlePath + ".map"
	require.FileExists(t, firstBundlePath)
	require.FileExists(t, firstSourceMapPath)

	require.NoError(t, os.WriteFile(helperFile, []byte(`export const message = "second build";`), 0644))

	result, err := ctx.compileTypescript("main.ts")
	require.NoError(t, err)
	require.Empty(t, result.Errors)

	secondHash, err := ctx.getSourceHash()
	require.NoError(t, err)
	require.NotEqual(t, firstHash, secondHash)

	secondBundlePath := filepath.Join(config.HiddenPath, fmt.Sprintf("bundle.%s.js", secondHash))
	secondSourceMapPath := secondBundlePath + ".map"

	require.NoFileExists(t, firstBundlePath)
	require.NoFileExists(t, firstSourceMapPath)
	require.FileExists(t, secondBundlePath)
	require.FileExists(t, secondSourceMapPath)

	bundleContent, err := os.ReadFile(secondBundlePath)
	require.NoError(t, err)
	require.Contains(t, string(bundleContent), "second build")
}

func setupLoaderTestProject(t *testing.T, folderName string) (*Context, *Config, string, string) {
	t.Helper()

	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, folderName)
	configFile := filepath.Join(tempDir, DefaultConfigFileName)

	require.NoError(t, Init(projectDir, configFile))

	config, err := LoadConfig(configFile)
	require.NoError(t, err)
	config.HiddenPath = filepath.Join(projectDir, DefaultHiddenFolderName)

	mainFile := filepath.Join(config.CCodePath, "main.ts")
	helperFile := filepath.Join(config.CCodePath, "helper.ts")

	return NewContext(config), config, mainFile, helperFile
}
