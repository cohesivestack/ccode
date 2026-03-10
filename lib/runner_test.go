package ccode

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_Run_ValidatesDefaultExportContextSignature(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestContext_Run_ValidatesDefaultExportContextSignature")
	processFile := filepath.Join(projectDir, "x", "generate.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte("export default function main() {}\n"), 0644))

	err := ctx.Run("x/generate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must export a default function with a single Context-typed parameter")
}

func TestContext_Run_ExecutesDefaultExport(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestContext_Run_ExecutesDefaultExport")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	helperFile := filepath.Join(projectDir, "x", "helper.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(helperFile, []byte(`export const message = "runner executed";`), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/types\";\nimport { message } from \"./helper\";\n\nexport default function main(ctx: Context) {\n\tctx.println(message);\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Equal(t, "runner executed\n", output.String())
}

func setupRunnerTestProject(t *testing.T, folderName string) (*Context, string) {
	t.Helper()

	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, folderName)
	configFile := filepath.Join(tempDir, DefaultConfigFileName)

	require.NoError(t, Init(projectDir, configFile))

	config, err := LoadConfig(configFile)
	require.NoError(t, err)

	return NewContext(config), config.Path
}
