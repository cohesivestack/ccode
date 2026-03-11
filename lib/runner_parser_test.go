package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_ParseJSONFromBytes(t *testing.T) {
	ctx := NewContext(nil)

	result, err := ctx.ParseJSONFromBytes([]byte(`{"name":"Carlos","count":2,"active":true}`))
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"name":   "Carlos",
		"count":  float64(2),
		"active": true,
	}, result)
}

func TestContext_ParseJSONFromBytes_ReturnsErrorForInvalidJSON(t *testing.T) {
	ctx := NewContext(nil)

	result, err := ctx.ParseJSONFromBytes([]byte(`{"name":`))
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestContext_ParseJSONFromString(t *testing.T) {
	ctx := NewContext(nil)

	result, err := ctx.ParseJSONFromString(`{"source":"string","items":["a","b"]}`)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"source": "string",
		"items":  []any{"a", "b"},
	}, result)
}

func TestContext_ParseJSONFromFile_UsesConfigCCodePath(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "data"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "ccode", "data"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "data", "input.json"), []byte(`{"source":"cwd"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "ccode", "data", "input.json"), []byte(`{"source":"ccode","enabled":true}`), 0644))

	config, err := NewConfig(&Config{
		CCodePath: "ccode",
	})
	require.NoError(t, err)

	ctx := NewContext(config)

	result, err := ctx.ParseJSONFromFile("data/input.json")
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"source":  "ccode",
		"enabled": true,
	}, result)
}

func TestContext_ParseJSONFromFile_ReturnsErrorForMissingFile(t *testing.T) {
	ctx := NewContext(&Config{
		CCodePath: t.TempDir(),
	})

	result, err := ctx.ParseJSONFromFile("missing.json")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file not found: missing.json")
}
