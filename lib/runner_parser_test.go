package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerContext_ParseJSONFromBytes(t *testing.T) {
	ctx := newRunnerJSONTestContext(t)

	value, err := ctx.ParseJSONFromBytes([]byte(`{"name":"Carlos","count":2,"active":true}`))
	require.NoError(t, err)

	object := value.ToObject(ctx.runtime)
	assert.Equal(t, []string{"name", "count", "active"}, object.Keys())
	assert.Equal(t, "Carlos", object.Get("name").String())
	assert.EqualValues(t, 2, object.Get("count").ToInteger())
	assert.True(t, object.Get("active").ToBoolean())
}

func TestRunnerContext_ParseJSONFromBytes_ReturnsErrorForInvalidJSON(t *testing.T) {
	ctx := newRunnerJSONTestContext(t)

	result, err := ctx.ParseJSONFromBytes([]byte(`{"name":`))
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unexpected end of JSON input")
}

func TestRunnerContext_ParseJSONFromString(t *testing.T) {
	ctx := newRunnerJSONTestContext(t)

	value, err := ctx.ParseJSONFromString(`{"source":"string","items":["a","b"]}`)
	require.NoError(t, err)

	object := value.ToObject(ctx.runtime)
	assert.Equal(t, []string{"source", "items"}, object.Keys())
	assert.Equal(t, "string", object.Get("source").String())
	items := object.Get("items").ToObject(ctx.runtime)
	assert.EqualValues(t, 2, items.Get("length").ToInteger())
	assert.Equal(t, "a", items.Get("0").String())
	assert.Equal(t, "b", items.Get("1").String())
}

func TestRunnerContext_ParseJSONFromFile_UsesConfigCCodePath(t *testing.T) {
	ctx := newRunnerJSONTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "data"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "data", "input.json"), []byte(`{"source":"ccode","enabled":true}`), 0644))

	value, err := ctx.ParseJSONFromFile("data/input.json")
	require.NoError(t, err)

	object := value.ToObject(ctx.runtime)
	assert.Equal(t, []string{"source", "enabled"}, object.Keys())
	assert.Equal(t, "ccode", object.Get("source").String())
	assert.True(t, object.Get("enabled").ToBoolean())
}

func TestRunnerContext_ParseJSONFromFile_ReturnsErrorForMissingFile(t *testing.T) {
	ctx := newRunnerJSONTestContext(t)

	result, err := ctx.ParseJSONFromFile("missing.json")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file not found: missing.json")
}

func TestRunnerContext_ParseJSONFromFile_PreservesNestedObjectOrder(t *testing.T) {
	ctx := newRunnerJSONTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "data"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "data", "ordered.json"), []byte("{\n  \"z\": 1,\n  \"a\": {\n    \"beta\": 1,\n    \"alpha\": 2\n  },\n  \"m\": 3\n}\n"), 0644))

	value, err := ctx.ParseJSONFromFile("data/ordered.json")
	require.NoError(t, err)

	object := value.ToObject(ctx.runtime)
	assert.Equal(t, []string{"z", "a", "m"}, object.Keys())
	assert.Equal(t, []string{"beta", "alpha"}, object.Get("a").ToObject(ctx.runtime).Keys())
}

func newRunnerJSONTestContext(t *testing.T) *RunnerContext {
	t.Helper()

	runtime := goja.New()
	ctx := &RunnerContext{
		ccodeContext: NewContext(&Config{CCodePath: t.TempDir()}),
		runtime:      runtime,
	}
	require.NoError(t, ctx.initializeJSONParser())
	return ctx
}
