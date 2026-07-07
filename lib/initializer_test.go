package ccode

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	templateassets "github.com/cohesivestack/ccode/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializer_InitCreatesProjectStructure(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	require.NoError(t, Init("cohesive", "configs/ccode.yaml", "v1.2.3"))

	projectPath := filepath.Join(tmp, "cohesive")
	configPath := filepath.Join(tmp, "configs", "ccode.yaml")
	hiddenPath := filepath.Join(projectPath, DefaultHiddenFolderName)
	buildPath := filepath.Join(hiddenPath, "build")
	contextPath := filepath.Join(hiddenPath, "lib", "context.ts")
	openapiPath := filepath.Join(hiddenPath, "lib", "openapi.ts")
	tsconfigPath := filepath.Join(projectPath, "tsconfig.json")

	require.DirExists(t, projectPath)
	require.DirExists(t, hiddenPath)
	require.DirExists(t, filepath.Join(hiddenPath, "lib"))
	require.DirExists(t, buildPath)
	require.FileExists(t, configPath)
	require.FileExists(t, contextPath)
	require.FileExists(t, openapiPath)
	require.FileExists(t, tsconfigPath)

	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "ccode_path: cohesive\nversion: v1.2.3\n", string(configContent))

	contextContent, err := os.ReadFile(contextPath)
	require.NoError(t, err)
	assert.Equal(t, templateassets.ContextTemplate, string(contextContent))

	openapiContent, err := os.ReadFile(openapiPath)
	require.NoError(t, err)
	assert.Equal(t, templateassets.OpenAPITemplate, string(openapiContent))

	tsconfigContent, err := os.ReadFile(tsconfigPath)
	require.NoError(t, err)
	assert.Equal(t, templateassets.TSConfigTemplate, string(tsconfigContent))
}

func TestInitializer_InitRefreshesGeneratedSupportFilesAndPreservesUserFiles(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	projectPath := filepath.Join(tmp, "ccode")
	hiddenLibPath := filepath.Join(projectPath, DefaultHiddenFolderName, "lib")
	configPath := filepath.Join(tmp, DefaultConfigFileName)
	contextPath := filepath.Join(hiddenLibPath, "context.ts")
	openapiPath := filepath.Join(hiddenLibPath, "openapi.ts")
	buildPath := filepath.Join(projectPath, DefaultHiddenFolderName, "build")
	buildCachePath := filepath.Join(buildPath, "process.old.js")
	statePath := filepath.Join(projectPath, DefaultHiddenFolderName, "state", "accelerators.json")
	tsconfigPath := filepath.Join(projectPath, "tsconfig.json")

	require.NoError(t, os.MkdirAll(hiddenLibPath, 0755))
	require.NoError(t, os.MkdirAll(buildPath, 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(statePath), 0755))
	require.NoError(t, os.WriteFile(configPath, []byte("ccode_path: ccode\nversion: v0.1.0\n"), 0644))
	require.NoError(t, os.WriteFile(contextPath, []byte("existing context"), 0644))
	require.NoError(t, os.WriteFile(openapiPath, []byte("existing openapi"), 0644))
	require.NoError(t, os.WriteFile(buildCachePath, []byte("cached bundle"), 0644))
	require.NoError(t, os.WriteFile(statePath, []byte(`{"version":1}`), 0644))
	require.NoError(t, os.WriteFile(tsconfigPath, []byte(`{"existing":true}`), 0644))

	var logs bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	require.NoError(t, Init("", "", "v1.2.3"))

	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "ccode_path: ccode\nversion: v1.2.3\n", string(configContent))

	contextContent, err := os.ReadFile(contextPath)
	require.NoError(t, err)
	assert.Equal(t, templateassets.ContextTemplate, string(contextContent))

	openapiContent, err := os.ReadFile(openapiPath)
	require.NoError(t, err)
	assert.Equal(t, templateassets.OpenAPITemplate, string(openapiContent))

	tsconfigContent, err := os.ReadFile(tsconfigPath)
	require.NoError(t, err)
	assert.Equal(t, `{"existing":true}`, string(tsconfigContent))

	assert.NoFileExists(t, buildCachePath)
	assert.FileExists(t, statePath)
	assert.Contains(t, logs.String(), "tsconfig already exists; not overwriting")
	require.DirExists(t, buildPath)
}

func TestInitializer_InitAddsVersionToExistingConfig(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	configPath := filepath.Join(tmp, DefaultConfigFileName)
	require.NoError(t, os.WriteFile(configPath, []byte("ccode_path: ccode\n"), 0644))

	require.NoError(t, Init("", "", "v1.2.3"))

	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "ccode_path: ccode\nversion: v1.2.3\n", string(configContent))
}
