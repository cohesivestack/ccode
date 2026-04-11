package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_GojaValueToTemplateDataPreservesObjectOrder(t *testing.T) {
	runtime := goja.New()

	value, err := runtime.RunString(`({ z: 1, nested: { beta: 1, alpha: 2 }, a: 3 })`)
	require.NoError(t, err)

	converted, err := gojaValueToTemplateData(value)
	require.NoError(t, err)

	dict, ok := converted.(exec.Dict)
	require.True(t, ok)
	require.Len(t, dict.Pairs, 3)
	assert.Equal(t, "z", dict.Pairs[0].Key.String())
	assert.Equal(t, "nested", dict.Pairs[1].Key.String())
	assert.Equal(t, "a", dict.Pairs[2].Key.String())

	nested := dict.Pairs[1].Value.Interface().(exec.Dict)
	require.Len(t, nested.Pairs, 2)
	assert.Equal(t, "beta", nested.Pairs[0].Key.String())
	assert.Equal(t, "alpha", nested.Pairs[1].Key.String())
}

func TestRenderer_RenderTemplateUsesConfigCCodePath(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "greeting.tpl"), []byte("Hello {{ data.name }}!"), 0644))

	value, err := ctx.runtime.RunString(`({ name: "Carlos" })`)
	require.NoError(t, err)

	rendered, err := ctx.RenderTemplate("templates/greeting.tpl", value)
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", rendered)
}

func newRunnerTemplateTestContext(t *testing.T) *RunnerContext {
	t.Helper()

	runtime := goja.New()
	rootDir := t.TempDir()
	ctx := &RunnerContext{
		ccodeContext: NewContext(&Config{
			CCodePath:  filepath.Join(rootDir, "ccode"),
			OutputPath: filepath.Join(rootDir, "output"),
		}),
		runtime: runtime,
	}
	require.NoError(t, ctx.initializeJSONParser())
	return ctx
}
