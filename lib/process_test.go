package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_ListProcessesRecursively(t *testing.T) {
	projectPath := t.TempDir()
	ctx := NewContext(&Config{
		CCodePath:  projectPath,
		HiddenPath: "internal-support",
	})

	paths := []string{
		"generate.process.ts",
		"api/client.process.ts",
		"api/nested/server.process.ts",
		"api/helper.ts",
		"templates/example.process.ts.tpl",
		".hidden/ignored.process.ts",
		"api/.cache/ignored.process.ts",
		"internal-support/ignored.process.ts",
	}
	for _, path := range paths {
		fullPath := filepath.Join(projectPath, filepath.FromSlash(path))
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte("export default function () {}\n"), 0644))
	}

	processes, err := ctx.ListProcesses()
	require.NoError(t, err)
	assert.Equal(t, []string{
		"api/client",
		"api/nested/server",
		"generate",
	}, processes)
}

func TestContext_ListProcessesReturnsErrorForMissingProjectPath(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "missing")
	ctx := NewContext(&Config{
		CCodePath:  projectPath,
		HiddenPath: ".ccode",
	})

	processes, err := ctx.ListProcesses()
	require.Error(t, err)
	assert.Nil(t, processes)
	assert.Contains(t, err.Error(), "list processes in")
}
