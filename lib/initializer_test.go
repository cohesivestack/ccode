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
	hiddenGitIgnorePath := filepath.Join(hiddenPath, ".gitignore")
	buildPath := filepath.Join(hiddenPath, "build")
	tsconfigPath := filepath.Join(projectPath, "tsconfig.json")

	require.DirExists(t, projectPath)
	require.DirExists(t, hiddenPath)
	require.DirExists(t, filepath.Join(hiddenPath, "lib"))
	require.DirExists(t, filepath.Join(hiddenPath, "lib", "openapi"))
	require.DirExists(t, filepath.Join(hiddenPath, "lib", "database"))
	require.DirExists(t, filepath.Join(hiddenPath, "lib", "typescript"))
	require.DirExists(t, buildPath)
	require.FileExists(t, configPath)
	require.FileExists(t, hiddenGitIgnorePath)
	require.FileExists(t, tsconfigPath)
	assertInstalledSupportFiles(t, filepath.Join(hiddenPath, "lib"))

	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "ccode_path: cohesive\nversion: v1.2.3\n", string(configContent))

	hiddenGitIgnoreContent, err := os.ReadFile(hiddenGitIgnorePath)
	require.NoError(t, err)
	assert.Equal(t, templateassets.HiddenGitIgnoreTemplate, string(hiddenGitIgnoreContent))
	assert.Contains(t, string(hiddenGitIgnoreContent), "!accelerators/**")

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
	hiddenPath := filepath.Join(projectPath, DefaultHiddenFolderName)
	hiddenLibPath := filepath.Join(hiddenPath, "lib")
	configPath := filepath.Join(tmp, DefaultConfigFileName)
	hiddenGitIgnorePath := filepath.Join(hiddenPath, ".gitignore")
	buildPath := filepath.Join(hiddenPath, "build")
	buildCachePath := filepath.Join(buildPath, "process.old.js")
	statePath := filepath.Join(hiddenPath, "accelerators", "generate-api", "handlers.go.accelerated.json")
	userSupportPath := filepath.Join(hiddenLibPath, "custom.ts")
	tsconfigPath := filepath.Join(projectPath, "tsconfig.json")

	require.NoError(t, os.MkdirAll(filepath.Join(hiddenLibPath, "openapi"), 0755))
	require.NoError(t, os.MkdirAll(buildPath, 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(statePath), 0755))
	require.NoError(t, os.WriteFile(configPath, []byte("ccode_path: ccode\nversion: v0.1.0\n"), 0644))
	require.NoError(t, os.WriteFile(hiddenGitIgnorePath, []byte("custom ignore"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(hiddenLibPath, "context.ts"), []byte("stale context"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(hiddenLibPath, "openapi", "index.ts"), []byte("stale openapi"), 0644))
	require.NoError(t, os.WriteFile(userSupportPath, []byte("user owned"), 0644))
	for _, name := range obsoleteSupportFileNames {
		require.NoError(t, os.WriteFile(filepath.Join(hiddenLibPath, name), []byte("obsolete generated support"), 0644))
	}
	require.NoError(t, os.WriteFile(buildCachePath, []byte("cached bundle"), 0644))
	require.NoError(t, os.WriteFile(statePath, []byte(`{"pending":true,"instructions":"","accelerated_checksum":"sha256:test","instructions_checksum":"","code":"dGVzdA=="}`+"\n"), 0644))
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
	assertInstalledSupportFiles(t, hiddenLibPath)

	for _, name := range obsoleteSupportFileNames {
		assert.NoFileExists(t, filepath.Join(hiddenLibPath, name))
	}
	userSupportContent, err := os.ReadFile(userSupportPath)
	require.NoError(t, err)
	assert.Equal(t, "user owned", string(userSupportContent))

	hiddenGitIgnoreContent, err := os.ReadFile(hiddenGitIgnorePath)
	require.NoError(t, err)
	assert.Equal(t, "custom ignore", string(hiddenGitIgnoreContent))

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

func assertInstalledSupportFiles(t *testing.T, hiddenLibPath string) {
	t.Helper()

	paths, err := templateassets.SupportFilePaths()
	require.NoError(t, err)
	assert.Equal(t, []string{
		"context.ts",
		"database/index.ts",
		"database/mariadb.ts",
		"database/mysql.ts",
		"database/postgresql.ts",
		"database/sqlite.ts",
		"go/index.ts",
		"internal/native.ts",
		"openapi/index.ts",
		"openapi/path.ts",
		"openapi/reference.ts",
		"openapi/types.ts",
		"openapi/v3_0.ts",
		"openapi/v3_1.ts",
		"openapi/v3_2.ts",
		"string/index.ts",
		"typescript/index.ts",
	}, paths)

	for _, path := range paths {
		embedded, err := templateassets.SupportFS.ReadFile(path)
		require.NoError(t, err)
		installed, err := os.ReadFile(filepath.Join(hiddenLibPath, filepath.FromSlash(path)))
		require.NoError(t, err)
		assert.Equal(t, embedded, installed, path)
	}
}
