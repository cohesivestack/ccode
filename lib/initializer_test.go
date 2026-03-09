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

func TestInit_CreatesProjectStructure(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	require.NoError(t, Init("cohesive", "configs/ccode.yaml"))

	projectPath := filepath.Join(tmp, "cohesive")
	configPath := filepath.Join(tmp, "configs", "ccode.yaml")
	hiddenPath := filepath.Join(projectPath, DefaultHiddenFolderName)
	buildPath := filepath.Join(hiddenPath, "build")
	typesPath := filepath.Join(hiddenPath, "lib", "types.ts")
	tsconfigPath := filepath.Join(projectPath, "tsconfig.json")

	require.DirExists(t, projectPath)
	require.DirExists(t, hiddenPath)
	require.DirExists(t, filepath.Join(hiddenPath, "lib"))
	require.DirExists(t, buildPath)
	require.FileExists(t, configPath)
	require.FileExists(t, typesPath)
	require.FileExists(t, tsconfigPath)

	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "path: cohesive\n", string(configContent))

	typesContent, err := os.ReadFile(typesPath)
	require.NoError(t, err)
	assert.Equal(t, templateassets.TypesTemplate, string(typesContent))

	tsconfigContent, err := os.ReadFile(tsconfigPath)
	require.NoError(t, err)
	assert.Equal(t, templateassets.TSConfigTemplate, string(tsconfigContent))
}

func TestInit_UsesDefaultsAndDoesNotOverwriteExistingFiles(t *testing.T) {
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
	typesPath := filepath.Join(hiddenLibPath, "types.ts")
	tsconfigPath := filepath.Join(projectPath, "tsconfig.json")

	require.NoError(t, os.MkdirAll(hiddenLibPath, 0755))
	require.NoError(t, os.WriteFile(configPath, []byte("path: existing"), 0644))
	require.NoError(t, os.WriteFile(typesPath, []byte("existing types"), 0644))
	require.NoError(t, os.WriteFile(tsconfigPath, []byte(`{"existing":true}`), 0644))

	var logs bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	require.NoError(t, Init("", ""))

	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "path: existing", string(configContent))

	typesContent, err := os.ReadFile(typesPath)
	require.NoError(t, err)
	assert.Equal(t, "existing types", string(typesContent))

	tsconfigContent, err := os.ReadFile(tsconfigPath)
	require.NoError(t, err)
	assert.Equal(t, `{"existing":true}`, string(tsconfigContent))

	assert.Contains(t, logs.String(), "config file already exists; not overwriting")
	assert.Contains(t, logs.String(), "types template already exists; not overwriting")
	assert.Contains(t, logs.String(), "tsconfig already exists; not overwriting")
	require.DirExists(t, filepath.Join(projectPath, DefaultHiddenFolderName, "build"))
}
