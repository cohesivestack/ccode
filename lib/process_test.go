package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_AddProcess(t *testing.T) {
	for _, process := range []string{"generator", "handlers/nested/generator", "handlers/generator.proc.ts"} {
		t.Run(process, func(t *testing.T) {
			projectPath := filepath.Join(t.TempDir(), "ccode")
			ctx := NewContext(&Config{CCodePath: projectPath})

			require.NoError(t, ctx.AddProcess(process))
			path := process
			if process != "handlers/generator.proc.ts" {
				path += ".proc.ts"
			}
			content, err := os.ReadFile(filepath.Join(projectPath, filepath.FromSlash(path)))
			require.NoError(t, err)
			assert.Equal(t, "import type { Context } from \"@ccode/context\";\n\nexport default function run(ctx: Context) {\n\n}\n", string(content))
		})
	}
}

func TestContext_AddProcessPreservesExistingFile(t *testing.T) {
	projectPath := t.TempDir()
	ctx := NewContext(&Config{CCodePath: projectPath})
	path := filepath.Join(projectPath, "generator.proc.ts")
	require.NoError(t, os.WriteFile(path, []byte("existing process"), 0644))

	err := ctx.AddProcess("generator")
	require.ErrorIs(t, err, os.ErrExist)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "existing process", string(content))
}

func TestContext_AddProcessRejectsInvalidPaths(t *testing.T) {
	for _, process := range []string{"", "   ", ".", "..", "../outside", "nested/../../outside", "/absolute/generator", "generator.ts", "generator.process.ts"} {
		t.Run(process, func(t *testing.T) {
			projectPath := filepath.Join(t.TempDir(), "ccode")
			ctx := NewContext(&Config{CCodePath: projectPath})

			require.Error(t, ctx.AddProcess(process))
			assert.NoDirExists(t, projectPath)
		})
	}
}

func TestContext_AddedProcessCanRun(t *testing.T) {
	ctx, _ := setupRunnerTestProject(t, "TestContext_AddedProcessCanRun")
	require.NoError(t, ctx.AddProcess("handlers/generator"))
	require.NoError(t, ctx.Run("handlers/generator"))
}

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
