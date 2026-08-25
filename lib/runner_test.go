package ccode

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_RunValidatesDefaultExportContextSignature(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunValidatesDefaultExportContextSignature")
	processFile := filepath.Join(projectDir, "x", "generate.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte("export default function main() {}\n"), 0644))

	err := ctx.Run("x/generate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must export a default function with a single Context-typed parameter")
}

func TestRunner_RunFailsWhenWorkspaceSupportFilesAreMissing(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunFailsWhenWorkspaceSupportFilesAreMissing")
	contextPath := filepath.Join(ctx.config.HiddenPath, "lib", "context.ts")
	require.NoError(t, os.Remove(contextPath))

	err := ctx.Run("x/generate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ccode workspace is not initialized or support files are missing")
	assert.Contains(t, err.Error(), "Run: ccode init")
	assert.Contains(t, err.Error(), filepath.Join(projectDir, DefaultHiddenFolderName, "lib", "context.ts"))
}

func TestRunner_RunRecreatesMissingBuildCache(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunRecreatesMissingBuildCache")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	buildPath := filepath.Join(ctx.config.HiddenPath, "build")

	require.NoError(t, os.RemoveAll(buildPath))
	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tctx.println(\"ok\");\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Equal(t, "ok\n", output.String())
	require.DirExists(t, buildPath)
}

func TestRunner_RunExecutesDefaultExport(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunExecutesDefaultExport")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	helperFile := filepath.Join(projectDir, "x", "helper.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(helperFile, []byte(`export const message = "runner executed";`), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\nimport { message } from \"./helper\";\n\nexport default function main(ctx: Context) {\n\tctx.println(message);\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Equal(t, "runner executed\n", output.String())
}

func TestRunner_RunRendersTemplates(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunRendersTemplates")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	templateFile := filepath.Join(projectDir, "templates", "greeting.tpl")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(templateFile), 0755))
	require.NoError(t, os.WriteFile(templateFile, []byte("Hello {{ data.name }}!"), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tconst rendered = ctx.renderTemplate(\"templates/greeting.tpl\", { name: \"Carlos\" });\n\tctx.generate(\"templates/greeting.tpl\", \"generated/greeting.txt\", { name: \"Carlos\" });\n\tctx.println(rendered);\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Equal(t, "Hello Carlos!\n", output.String())

	renderedFilePath := filepath.Join(filepath.Dir(projectDir), "output", "generated", "greeting.txt")
	content, err := os.ReadFile(renderedFilePath)
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", string(content))
}

func TestRunner_RunRendersTemplatesWithDeterministicObjectOrder(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunRendersTemplatesWithDeterministicObjectOrder")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	templateFile := filepath.Join(projectDir, "templates", "ordered.tpl")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(templateFile), 0755))
	require.NoError(t, os.WriteFile(templateFile, []byte("{% for key, value in data %}{{ key }} {% endfor %}|{% for key, value in data.nested %}{{ key }} {% endfor %}"), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tconst model = { z: 1, nested: { beta: 1, alpha: 2 }, a: 3 };\n\tconst rendered = ctx.renderTemplate(\"templates/ordered.tpl\", model);\n\tctx.generate(\"templates/ordered.tpl\", \"generated/ordered.txt\", model);\n\tctx.println(rendered);\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Equal(t, "z nested a |beta alpha \n", output.String())

	renderedFilePath := filepath.Join(filepath.Dir(projectDir), "output", "generated", "ordered.txt")
	content, err := os.ReadFile(renderedFilePath)
	require.NoError(t, err)
	assert.Equal(t, "z nested a |beta alpha ", string(content))
}

func TestRunner_RunTemplateErrorsCanBeCaughtInTypescript(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunTemplateErrorsCanBeCaughtInTypescript")
	processFile := filepath.Join(projectDir, "x", "generate.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\ttry {\n\t\tctx.renderTemplate(\"templates/missing.tpl\", { name: \"Carlos\" });\n\t} catch (e: any) {\n\t\tif (!(e instanceof GoError)) {\n\t\t\tthrow e;\n\t\t}\n\t\tctx.println(e.value.Error());\n\t}\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Contains(t, output.String(), "parse template \"templates/missing.tpl\"")
}

func TestRunner_RunParsesJSONThroughRunnerContext(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunParsesJSONThroughRunnerContext")
	processFile := filepath.Join(projectDir, "x", "generate.ts")

	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tempRoot := filepath.Dir(projectDir)
	require.NoError(t, os.Chdir(tempRoot))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	require.NoError(t, os.MkdirAll(filepath.Join(tempRoot, "data"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "data"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tempRoot, "data", "input.json"), []byte(`{"source":"cwd"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "data", "input.json"), []byte(`{"source":"file","enabled":true}`), 0644))
	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tconst bytes = '{\"source\":\"bytes\",\"count\":2}'.split('').map((char) => char.charCodeAt(0));\n\tconst fromBytes = ctx.parseJSONFromBytes(bytes);\n\tconst fromString = ctx.parseJSONFromString('{\"source\":\"string\",\"items\":[\"a\",\"b\"]}');\n\tconst fromFile = ctx.parseJSONFromFile('data/input.json');\n\tctx.println(JSON.stringify({ fromBytes, fromString, fromFile }));\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.JSONEq(t, `{"fromBytes":{"source":"bytes","count":2},"fromString":{"source":"string","items":["a","b"]},"fromFile":{"source":"file","enabled":true}}`, output.String())
}

func TestRunner_RunParseJSONPreservesDeterministicObjectKeyOrder(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunParseJSONPreservesDeterministicObjectKeyOrder")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	dataFile := filepath.Join(projectDir, "data", "input.json")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(dataFile), 0755))
	require.NoError(t, os.WriteFile(dataFile, []byte("{\n  \"z\": 1,\n  \"a\": {\n    \"beta\": 1,\n    \"alpha\": 2\n  },\n  \"m\": 3\n}\n"), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tconst model = ctx.parseJSONFromFile(\"data/input.json\");\n\tctx.println(JSON.stringify({\n\t\trootKeys: Object.keys(model),\n\t\tnestedKeys: Object.keys(model.a),\n\t}));\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.JSONEq(t, `{"rootKeys":["z","a","m"],"nestedKeys":["beta","alpha"]}`, output.String())
}

func TestRunner_RunParseJSONErrorsCanBeCaughtInTypescript(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunParseJSONErrorsCanBeCaughtInTypescript")
	processFile := filepath.Join(projectDir, "x", "generate.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tfor (const run of [\n\t\t() => ctx.parseJSONFromBytes([123]),\n\t\t() => ctx.parseJSONFromString('{'),\n\t\t() => ctx.parseJSONFromFile('missing.json'),\n\t]) {\n\t\ttry {\n\t\t\trun();\n\t\t} catch (e: any) {\n\t\t\tif (!(e instanceof GoError)) {\n\t\t\t\tthrow e;\n\t\t\t}\n\t\t\tctx.println(e.value.Error());\n\t\t}\n\t}\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Contains(t, output.String(), "Unexpected end of JSON input")
	assert.Contains(t, output.String(), "file not found: missing.json")
}

func TestRunner_RunParsesOpenAPIThroughRunnerContextWithDeterministicOrder(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunParsesOpenAPIThroughRunnerContextWithDeterministicOrder")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	specFile := filepath.Join(projectDir, "specs", "api.yaml")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(specFile), 0755))
	require.NoError(t, os.WriteFile(specFile, []byte(testOpenAPI3Document), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tconst spec = ctx.parseOpenAPIFromFile(\"specs/api.yaml\");\n\tconst pathKeys = Object.keys(spec.paths ?? {});\n\tconst sample = spec.components?.schemas?.Sample;\n\tif (!sample || sample === true || sample === false || \"$ref\" in sample) {\n\t\tthrow new Error(\"unexpected schema shape\");\n\t}\n\tconst propertyKeys = Object.keys(sample.properties ?? {});\n\tctx.println(JSON.stringify({ pathKeys, propertyKeys }));\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.JSONEq(t, `{"pathKeys":["/z","/a","/m"],"propertyKeys":["beta","alpha"]}`, output.String())
}

func TestRunner_RunParsesOpenAPIWithExpectedVersion(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunParsesOpenAPIWithExpectedVersion")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	specFile := filepath.Join(projectDir, "specs", "api.yaml")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(specFile), 0755))
	require.NoError(t, os.WriteFile(specFile, []byte(testOpenAPI3Document), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tconst spec = ctx.parseOpenAPIFromFile(\"specs/api.yaml\", { expectedVersion: \"3.1\" });\n\tctx.println(spec.openapi);\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Equal(t, "3.1.0\n", output.String())
}

func TestRunner_RunParsesResolvedExternalOpenAPIPathItemWithProvenance(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunParsesResolvedExternalOpenAPIPathItemWithProvenance")
	processFile := filepath.Join(projectDir, "x", "generate.ts")
	specFile := filepath.Join(projectDir, "app", "server.yaml")
	pathFile := filepath.Join(projectDir, "app", "paths", "countries.yaml")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(pathFile), 0755))
	require.NoError(t, os.WriteFile(specFile, []byte(testOpenAPIExternalPathDocument), 0644))
	require.NoError(t, os.WriteFile(pathFile, []byte("countries:\n  get:\n    operationId: getCountries\n    responses:\n      '200':\n        description: OK\n"), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte(`import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  const spec = ctx.parseOpenAPIFromFile("app/server.yaml", {
    expectedVersion: "3.1",
  });
  const pathItem = spec.paths?.["/countries"];
  if (!pathItem?.$ref || !pathItem.get?.operationId) {
    throw new Error("expected a resolved Path Item with provenance");
  }
  const sourceFile = pathItem.$ref.split("#", 1)[0].split("/").at(-1);
  ctx.println(JSON.stringify({ sourceFile, operationId: pathItem.get.operationId }));
}
`), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.JSONEq(t, `{"sourceFile":"countries.yaml","operationId":"getCountries"}`, output.String())
}

func TestRunner_RunOpenAPIErrorsCanBeCaughtInTypescript(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_RunOpenAPIErrorsCanBeCaughtInTypescript")
	processFile := filepath.Join(projectDir, "x", "generate.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tfor (const run of [\n\t\t() => ctx.parseOpenAPIFromString(\"swagger: '2.0'\\ninfo:\\n  title: Legacy\\n  version: 1.0.0\\npaths: {}\\n\"),\n\t\t() => ctx.parseOpenAPIFromFile(\"missing.yaml\"),\n\t]) {\n\t\ttry {\n\t\t\trun();\n\t\t} catch (e: any) {\n\t\t\tif (!(e instanceof GoError)) {\n\t\t\t\tthrow e;\n\t\t\t}\n\t\t\tctx.println(e.value.Error());\n\t\t}\n\t}\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate"))
	assert.Contains(t, output.String(), "swagger is not supported")
	assert.Contains(t, output.String(), "file not found: missing.yaml")
}

func setupRunnerTestProject(t *testing.T, folderName string) (*Context, string) {
	t.Helper()

	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, folderName)
	configFile := filepath.Join(tempDir, DefaultConfigFileName)

	require.NoError(t, Init(projectDir, configFile, "v1.2.3"))

	config, err := LoadConfig(configFile)
	require.NoError(t, err)
	config.OutputPath = filepath.Join(tempDir, "output")

	return NewContext(config), config.CCodePath
}
