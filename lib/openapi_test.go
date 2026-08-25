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

func TestOpenAPI_ParseOpenAPIFromFileAcceptsExpectedVersion(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs", "api.yaml"), []byte(testOpenAPI3Document), 0644))

	value, err := ctx.ParseOpenAPIFromFile(
		"specs/api.yaml",
		ctx.runtime.ToValue(map[string]any{"expectedVersion": "3.1"}),
	)
	require.NoError(t, err)

	document := value.ToObject(ctx.runtime)
	assert.Equal(t, "3.1.0", document.Get("openapi").String())
}

func TestOpenAPI_ParseOpenAPIFromFileAcceptsOpenAPI32(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs", "api.yaml"), []byte(testOpenAPI32Document), 0644))

	value, err := ctx.ParseOpenAPIFromFile(
		"specs/api.yaml",
		ctx.runtime.ToValue(map[string]any{"expectedVersion": "3.2"}),
	)
	require.NoError(t, err)

	document := value.ToObject(ctx.runtime)
	assert.Equal(t, "3.2.0", document.Get("openapi").String())
	assert.Equal(t, "https://example.com/openapi.yaml", document.Get("$self").String())
	queryOperation := document.Get("paths").ToObject(ctx.runtime).Get("/pets").ToObject(ctx.runtime).Get("query")
	assert.False(t, goja.IsUndefined(queryOperation))
	mediaTypes := document.Get("components").ToObject(ctx.runtime).Get("mediaTypes").ToObject(ctx.runtime)
	itemSchema := mediaTypes.Get("EventStream").ToObject(ctx.runtime).Get("itemSchema").ToObject(ctx.runtime)
	assert.Equal(t, "string", itemSchema.Get("type").String())
}

func TestOpenAPI_ParseOpenAPIFromFileRejectsUnexpectedVersion(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "specs", "api.yaml"), []byte(testOpenAPI3Document), 0644))

	result, err := ctx.ParseOpenAPIFromFile(
		"specs/api.yaml",
		ctx.runtime.ToValue(map[string]any{"expectedVersion": "3.0"}),
	)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "expected OpenAPI 3.0.x, but specs/api.yaml declares 3.1.0")
}

func TestOpenAPI_ParseOpenAPIFromFileRejectsUnsupportedExpectedVersion(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)

	result, err := ctx.ParseOpenAPIFromFile(
		"specs/api.yaml",
		ctx.runtime.ToValue(map[string]any{"expectedVersion": "2.0"}),
	)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), `unsupported expected OpenAPI version "2.0"`)
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

func TestOpenAPI_ParseOpenAPIFromFileResolvesExternalPathItemFragmentWithProvenance(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "app/server.yaml", `openapi: 3.1.0
info:
  title: Example API
  version: 0.1.0
paths:
  /countries:
    $ref: ./paths/countries.yaml#/countries
`)
	writeOpenAPITestFile(t, ctx, "app/paths/countries.yaml", `countries:
  get:
    operationId: getCountries
    responses:
      '200':
        description: OK
`)

	value, err := ctx.ParseOpenAPIFromFile("app/server.yaml", ctx.runtime.ToValue(map[string]any{"expectedVersion": "3.1"}))
	require.NoError(t, err)

	pathItem := openAPITestPathItem(t, ctx, value, "/countries")
	assert.Equal(t, "./paths/countries.yaml#/countries", pathItem.Get("$ref").String())
	assert.Equal(t, "getCountries", pathItem.Get("get").ToObject(ctx.runtime).Get("operationId").String())
}

func TestOpenAPI_ParseOpenAPIFromFileResolvesReferencedPathItemWithMultipleMethods(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "server.yaml", testOpenAPIExternalPathDocument)
	writeOpenAPITestFile(t, ctx, "paths/countries.yaml", `countries:
  get:
    operationId: getCountries
    responses:
      '200':
        description: OK
  post:
    operationId: createCountry
    responses:
      '201':
        description: Created
`)

	value, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.NoError(t, err)

	pathItem := openAPITestPathItem(t, ctx, value, "/countries")
	assert.Equal(t, "getCountries", pathItem.Get("get").ToObject(ctx.runtime).Get("operationId").String())
	assert.Equal(t, "createCountry", pathItem.Get("post").ToObject(ctx.runtime).Get("operationId").String())
}

func TestOpenAPI_ParseOpenAPIFromFileResolvesNestedExternalSchemaReferences(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "server.yaml", testOpenAPIExternalPathDocument)
	writeOpenAPITestFile(t, ctx, "paths/countries.yaml", `countries:
  get:
    operationId: getCountries
    responses:
      '200':
        description: OK
        content:
          application/json:
            schema:
              $ref: ../schemas/countries.yaml#/Countries
`)
	writeOpenAPITestFile(t, ctx, "schemas/countries.yaml", `Countries:
  type: array
  items:
    $ref: ./country.yaml#/Country
`)
	writeOpenAPITestFile(t, ctx, "schemas/country.yaml", `Country:
  type: object
  properties:
    code:
      type: string
`)

	value, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.NoError(t, err)

	pathItem := openAPITestPathItem(t, ctx, value, "/countries")
	schema := pathItem.Get("get").ToObject(ctx.runtime).
		Get("responses").ToObject(ctx.runtime).
		Get("200").ToObject(ctx.runtime).
		Get("content").ToObject(ctx.runtime).
		Get("application/json").ToObject(ctx.runtime).
		Get("schema").ToObject(ctx.runtime)
	assert.Equal(t, "../schemas/countries.yaml#/Countries", schema.Get("$ref").String())
	assert.Equal(t, "array", schema.Get("type").String())
	items := schema.Get("items").ToObject(ctx.runtime)
	assert.Equal(t, "./country.yaml#/Country", items.Get("$ref").String())
	assert.Equal(t, "object", items.Get("type").String())
	assert.Equal(t, "string", items.Get("properties").ToObject(ctx.runtime).Get("code").ToObject(ctx.runtime).Get("type").String())
}

func TestOpenAPI_ParseOpenAPIFromFileResolvesInternalReferences(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "server.yaml", `openapi: 3.1.0
info:
  title: Internal API
  version: 1.0.0
paths:
  /health:
    $ref: '#/components/pathItems/Health'
components:
  pathItems:
    Health:
      get:
        operationId: getHealth
        responses:
          '200':
            description: OK
`)

	value, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.NoError(t, err)

	pathItem := openAPITestPathItem(t, ctx, value, "/health")
	assert.Equal(t, "#/components/pathItems/Health", pathItem.Get("$ref").String())
	assert.Equal(t, "getHealth", pathItem.Get("get").ToObject(ctx.runtime).Get("operationId").String())
}

func TestOpenAPI_ParseOpenAPIFromFileResolvesEscapedJSONPointerTokens(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "server.yaml", `openapi: 3.1.0
info:
  title: Escaped Pointer API
  version: 1.0.0
paths:
  /escaped:
    $ref: './paths.yaml#/a~1b~0c'
`)
	writeOpenAPITestFile(t, ctx, "paths.yaml", `'a/b~c':
  get:
    operationId: escapedPointer
    responses:
      '200':
        description: OK
`)

	value, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.NoError(t, err)
	pathItem := openAPITestPathItem(t, ctx, value, "/escaped")
	assert.Equal(t, "escapedPointer", pathItem.Get("get").ToObject(ctx.runtime).Get("operationId").String())
}

func TestOpenAPI_ParseOpenAPIFromFileReturnsContextForMissingReferencedFile(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "server.yaml", testOpenAPIExternalPathDocument)

	value, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.Error(t, err)
	assert.Nil(t, value)
	assert.Contains(t, err.Error(), `resolve reference "./paths/countries.yaml#/countries"`)
	assert.Contains(t, err.Error(), "read referenced OpenAPI file")
	assert.Contains(t, err.Error(), "countries.yaml")
}

func TestOpenAPI_ParseOpenAPIFromFileReturnsContextForMissingFragment(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "server.yaml", testOpenAPIExternalPathDocument)
	writeOpenAPITestFile(t, ctx, "paths/countries.yaml", "anotherPath: {}\n")

	value, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.Error(t, err)
	assert.Nil(t, value)
	assert.Contains(t, err.Error(), "JSON Pointer #/countries does not exist")
}

func TestOpenAPI_ParseOpenAPIFromFileDetectsCircularReferences(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "server.yaml", `openapi: 3.1.0
info:
  title: Circular API
  version: 1.0.0
paths: {}
components:
  schemas:
    Node:
      $ref: ./schemas/node.yaml#/Node
`)
	writeOpenAPITestFile(t, ctx, "schemas/node.yaml", `Node:
  type: object
  properties:
    child:
      $ref: '#/Node'
`)

	value, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.NoError(t, err)

	node := value.ToObject(ctx.runtime).Get("components").ToObject(ctx.runtime).Get("schemas").ToObject(ctx.runtime).Get("Node").ToObject(ctx.runtime)
	assert.Equal(t, "./schemas/node.yaml#/Node", node.Get("$ref").String())
	child := node.Get("properties").ToObject(ctx.runtime).Get("child").ToObject(ctx.runtime)
	assert.Equal(t, "#/Node", child.Get("$ref").String())
}

func TestOpenAPI_ParseOpenAPIFromFilePreservesDeterministicOrderAfterResolution(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "server.yaml", testOpenAPIExternalPathDocument)
	writeOpenAPITestFile(t, ctx, "paths/countries.yaml", `countries:
  summary: Countries
  get:
    operationId: getCountries
    responses:
      '200':
        description: OK
  post:
    operationId: createCountry
    responses:
      '201':
        description: Created
`)

	first, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.NoError(t, err)
	second, err := ctx.ParseOpenAPIFromFile("server.yaml")
	require.NoError(t, err)

	firstKeys := openAPITestPathItem(t, ctx, first, "/countries").Keys()
	secondKeys := openAPITestPathItem(t, ctx, second, "/countries").Keys()
	assert.Equal(t, []string{"$ref", "summary", "get", "post"}, firstKeys)
	assert.Equal(t, firstKeys, secondKeys)
}

func writeOpenAPITestFile(t *testing.T, ctx *RunnerContext, relativePath, contents string) {
	t.Helper()
	fullPath := filepath.Join(ctx.ccodeContext.config.CCodePath, relativePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
	require.NoError(t, os.WriteFile(fullPath, []byte(contents), 0644))
}

func openAPITestPathItem(t *testing.T, ctx *RunnerContext, value goja.Value, path string) *goja.Object {
	t.Helper()
	pathItem := value.ToObject(ctx.runtime).Get("paths").ToObject(ctx.runtime).Get(path)
	require.False(t, goja.IsUndefined(pathItem))
	return pathItem.ToObject(ctx.runtime)
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

const testOpenAPI32Document = `openapi: 3.2.0
$self: https://example.com/openapi.yaml
info:
  title: OpenAPI 3.2 API
  version: 1.0.0
paths:
  /pets:
    query:
      responses:
        '200':
          summary: Stream opened
components:
  mediaTypes:
    EventStream:
      itemSchema:
        type: string
`

const testOpenAPIExternalPathDocument = `openapi: 3.1.0
info:
  title: Example API
  version: 0.1.0
paths:
  /countries:
    $ref: ./paths/countries.yaml#/countries
`
