package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPI_ParseOpenAPIFromBytes(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	value, err := ctx.ParseOpenAPIFromBytes([]byte(testOpenAPI3Document))
	require.NoError(t, err)

	document := value.ToObject(ctx.runtime)
	assert.Equal(t, "3.1.0", document.Get("openapi").String())
	assert.Equal(t, []string{"/z", "/a", "/m"}, document.Get("paths").ToObject(ctx.runtime).Keys())
}

func TestOpenAPI_ParseOpenAPIFromString(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	value, err := ctx.ParseOpenAPIFromString(testOpenAPI3Document)
	require.NoError(t, err)

	document := value.ToObject(ctx.runtime)
	assert.Equal(t, "Ordered API", document.Get("info").ToObject(ctx.runtime).Get("title").String())
	assert.Equal(t, []string{"beta", "alpha"}, document.Get("components").ToObject(ctx.runtime).Get("schemas").ToObject(ctx.runtime).Get("Sample").ToObject(ctx.runtime).Get("properties").ToObject(ctx.runtime).Keys())
}

func TestOpenAPI_ParseOpenAPIFromFileUsesConfigCCodePath(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs", "api.yaml"), []byte(testOpenAPI3Document), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(t.TempDir(), "api.yaml"), []byte("openapi: 3.1.0\ninfo:\n  title: Wrong\n  version: 1.0.0\npaths: {}\n"), 0644))

	value, err := ctx.ParseOpenAPIFromFile("specs/api.yaml")
	require.NoError(t, err)

	document := value.ToObject(ctx.runtime)
	assert.Equal(t, "Ordered API", document.Get("info").ToObject(ctx.runtime).Get("title").String())
	assert.Equal(t, []string{"/z", "/a", "/m"}, document.Get("paths").ToObject(ctx.runtime).Keys())
}

func TestOpenAPI_ParseOpenAPIFromFileReturnsErrorForMissingFile(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	result, err := ctx.ParseOpenAPIFromFile("missing.yaml")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file not found: missing.yaml")
}

func TestOpenAPI_ParseOpenAPIFromStringReturnsErrorForSwaggerV2(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	result, err := ctx.ParseOpenAPIFromString("swagger: '2.0'\ninfo:\n  title: Legacy\n  version: 1.0.0\npaths: {}\n")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "swagger is not supported")
}

func TestOpenAPI_ParseOpenAPIFromFilePreservesDeterministicOrder(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs", "ordered.yaml"), []byte(testOpenAPI3Document), 0644))

	value, err := ctx.ParseOpenAPIFromFile("specs/ordered.yaml")
	require.NoError(t, err)

	document := value.ToObject(ctx.runtime)
	assert.Equal(t, []string{"/z", "/a", "/m"}, document.Get("paths").ToObject(ctx.runtime).Keys())
	assert.Equal(t, []string{"beta", "alpha"}, document.Get("components").ToObject(ctx.runtime).Get("schemas").ToObject(ctx.runtime).Get("Sample").ToObject(ctx.runtime).Get("properties").ToObject(ctx.runtime).Keys())
}

func newRunnerOpenAPITestContext(t *testing.T) *RunnerContext {
	t.Helper()

	runtime := goja.New()
	ctx := &RunnerContext{
		ccodeContext: NewContext(&Config{CCodePath: t.TempDir()}),
		runtime:      runtime,
	}
	require.NoError(t, ctx.initializeJSONParser())
	return ctx
}

const testOpenAPI3Document = `openapi: 3.1.0
info:
  title: Ordered API
  version: 1.0.0
paths:
  /z:
    get:
      responses:
        '200':
          description: ok
  /a:
    get:
      responses:
        '200':
          description: ok
  /m:
    get:
      responses:
        '200':
          description: ok
components:
  schemas:
    Sample:
      type: object
      properties:
        beta:
          type: string
        alpha:
          type: string
`
