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
		"generate.proc.ts",
		"api/client.proc.ts",
		"api/nested/server.proc.ts",
		"api/helper.ts",
		"api/legacy.process.ts",
		"templates/example.proc.ts.tpl",
		".hidden/ignored.proc.ts",
		"api/.cache/ignored.proc.ts",
		"internal-support/ignored.proc.ts",
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
