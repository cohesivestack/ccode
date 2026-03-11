package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_TemplateToString_UsesConfigCCodePath(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "ccode", "templates"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "templates", "greeting.tpl"), []byte("wrong template"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "ccode", "templates", "greeting.tpl"), []byte("Hello {{ data.name }}!"), 0644))

	config, err := NewConfig(&Config{
		CCodePath: "ccode",
	})
	require.NoError(t, err)

	ctx := NewContext(config)

	result, err := ctx.TemplateToString("templates/greeting.tpl", map[string]any{
		"name": "Carlos",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", result)
}

func TestContext_TemplateToFile_UsesConfigOutputPath(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "ccode", "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "nested"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "generated", "nested"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "ccode", "templates", "greeting.tpl"), []byte("Hello {{ data.name }}!"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "nested", "greeting.txt"), []byte("wrong output location"), 0644))

	config, err := NewConfig(&Config{
		CCodePath:  "ccode",
		OutputPath: filepath.Join(tmp, "generated"),
	})
	require.NoError(t, err)

	ctx := NewContext(config)

	err = ctx.TemplateToFile("templates/greeting.tpl", "nested/greeting.txt", map[string]any{
		"name": "Carlos",
	})
	require.NoError(t, err)

	outputFilePath := filepath.Join(tmp, "generated", "nested", "greeting.txt")
	require.FileExists(t, outputFilePath)

	content, err := os.ReadFile(outputFilePath)
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", string(content))

	wrongLocationContent, err := os.ReadFile(filepath.Join(tmp, "nested", "greeting.txt"))
	require.NoError(t, err)
	assert.Equal(t, "wrong output location", string(wrongLocationContent))
}
