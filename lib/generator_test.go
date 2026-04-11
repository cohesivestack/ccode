package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateUsesConfigOutputPath(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "greeting.tpl"), []byte("Hello {{ data.name }}!"), 0644))

	value, err := ctx.runtime.RunString(`({ name: "Carlos" })`)
	require.NoError(t, err)

	err = ctx.Generate("templates/greeting.tpl", "nested/greeting.txt", value)
	require.NoError(t, err)

	outputFilePath := filepath.Join(ctx.ccodeContext.config.OutputPath, "nested", "greeting.txt")
	require.FileExists(t, outputFilePath)

	content, err := os.ReadFile(outputFilePath)
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", string(content))
}

func TestGenerator_GenerateReturnsErrorForBlankFilePath(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "greeting.tpl"), []byte("Hello {{ data.name }}!"), 0644))

	value, err := ctx.runtime.RunString(`({ name: "Carlos" })`)
	require.NoError(t, err)

	err = ctx.Generate("templates/greeting.tpl", " ", value)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file path is required")
}

func TestGenerator_GenerateWritesToAbsolutePath(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "greeting.tpl"), []byte("Hello {{ data.name }}!"), 0644))

	value, err := ctx.runtime.RunString(`({ name: "Carlos" })`)
	require.NoError(t, err)

	absoluteOutputPath := filepath.Join(t.TempDir(), "absolute", "greeting.txt")
	err = ctx.Generate("templates/greeting.tpl", absoluteOutputPath, value)
	require.NoError(t, err)

	require.FileExists(t, absoluteOutputPath)
	content, err := os.ReadFile(absoluteOutputPath)
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", string(content))

	assert.NoFileExists(t, filepath.Join(ctx.ccodeContext.config.OutputPath, "absolute", "greeting.txt"))
}
