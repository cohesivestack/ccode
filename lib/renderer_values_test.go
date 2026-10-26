package ccode

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderGojaTemplate(t *testing.T, script, template string) (string, error) {
	t.Helper()
	ctx := newRunnerTemplateTestContext(t)
	value, err := ctx.runtime.RunString(script)
	require.NoError(t, err)
	writeRendererTestTemplate(t, ctx, template)
	return ctx.RenderTemplate("templates/filter.tpl", value)
}

func TestRenderer_TemplateBridgePreservesCyclesAndSharedObjects(t *testing.T) {
	runtime := goja.New()
	value, err := runtime.RunString(`(() => {
   const root = {name: "root"};
   const child = {name: "child", parent: root};
   root.first = child;
   root.second = child;
   root.self = root;
   root.children = [child];
   root.children.push(root.children);
   return root;
 })()`)
	require.NoError(t, err)
	converted, err := gojaValueToTemplateData(runtime, value)
	require.NoError(t, err)
	root := exec.AsValue(converted)
	self, ok := root.GetItem("self")
	require.True(t, ok)
	// Ordered dictionary values share their backing pairs, including cycles.
	rootDict := converted.(exec.Dict)
	selfDict := self.Interface().(exec.Dict)
	require.Same(t, &rootDict.Pairs[0], &selfDict.Pairs[0])
	first, ok := root.GetItem("first")
	require.True(t, ok)
	second, ok := root.GetItem("second")
	require.True(t, ok)
	firstDict := first.Interface().(exec.Dict)
	secondDict := second.Interface().(exec.Dict)
	require.Same(t, &firstDict.Pairs[0], &secondDict.Pairs[0])
	parent, ok := first.GetItem("parent")
	require.True(t, ok)
	parentDict := parent.Interface().(exec.Dict)
	require.Same(t, &rootDict.Pairs[0], &parentDict.Pairs[0])
	children, ok := root.GetItem("children")
	require.True(t, ok)
	array := children.Interface().([]any)
	require.Same(t, &array[0], &array[1].([]any)[0])
}

func TestRenderer_TemplateBridgeCallsKeepTheirReceiver(t *testing.T) {
	rendered, err := renderGojaTemplate(t, `(() => {
  function label(suffix) { return this.prefix + suffix; }
  return {left: {prefix: "L", label}, right: {prefix: "R", label}};
 })()`, `{% set detached = data.left.label %}{{ detached("!") }}|{{ data.right.label("?") }}`)
	require.NoError(t, err)
	assert.Equal(t, "L!|R?", rendered)
}

func TestRenderer_TemplateBridgeConvertsOptionsAndResults(t *testing.T) {
	rendered, err := renderGojaTemplate(t, `({
   read(options) {
     if (options.enabled !== true || options.none !== null || options.nested.values[1] !== 2 ||
         options.nested.values[2].text !== "ok" || !Object.prototype.hasOwnProperty.call(options, "__proto__")) {
       throw new Error("incorrect options conversion");
     }
     const result = {label: options.region, parent: this};
     result.self = result;
     return result;
   },
   missing() { return undefined; },
   name: "original"
 })`, `{% set result = data.read({"region": "eu", "enabled": true, "none": none, "nested": {"values": [1, 2, {"text": "ok"}]}, "__proto__": "data"}) %}{{ result.label }}|{{ result.self.label }}|{{ result.parent.name }}|{{ data.missing() is none }}`)
	require.NoError(t, err)
	assert.Equal(t, "eu|eu|original|True", rendered)
}

func TestRenderer_TemplateBridgePassesOriginalObjectsToMethods(t *testing.T) {
	rendered, err := renderGojaTemplate(t, `(() => {
  const data = {
    child: {}, other: {}, empty: [], otherList: [], items: [1, 2],
    count(items) { return items.length; },
    same(value) { return value === this.child; },
    nested(options) { return options.child === this.child; },
    distinguish(a,b,c,d) { return a === this.child && b === this.other && a !== b && c === this.empty && d === this.otherList && c !== d; }
  };
  data.child.parent = data;
  return data;
 })()`, `{{ data.same(data.child) }}|{{ data.nested({"child": data.child}) }}|{{ data.distinguish(data.child, data.other, data.empty, data.otherList) }}|{{ data.count(data.items[:1]) }}|{{ data.count(data.items[:0]) }}`)
	require.NoError(t, err)
	assert.Equal(t, "True|True|True|1|0", rendered)
}

func TestRenderer_TemplateBridgePreservesRawDictionaryBehavior(t *testing.T) {
	rendered, err := renderGojaTemplate(t, `({z: {name: "last"}, a: {name: "first"}})`, `{% for key, value in data %}{{ key }}={{ value.name }};{% endfor %}|{{ data["a"].name }}|{{ data | length }}`)
	require.NoError(t, err)
	assert.Equal(t, "z=last;a=first;|first|2", rendered)
}

func TestRenderer_TemplateBridgeReportsCallAndGetterErrors(t *testing.T) {
	for _, test := range []struct{ name, script, template, message string }{
		{"exception", `({fail() {throw new Error("method exploded");}})`, `{{ data.fail() }}`, "method exploded"},
		{"keyword arguments", `({read(value) {return value;}})`, `{{ data.read(value="x") }}`, "positional arguments only"},
		{"getter", `({get broken() {throw new Error("getter exploded");}})`, `{{ data.broken }}`, "getter exploded"},
		{"result getter", `({read() {return {get broken() {throw new Error("result getter exploded");}};}})`, `{{ data.read().broken }}`, "result getter exploded"},
		{"non-string dictionary key", `({read(value) {return value;}})`, `{{ data.read({1: "value"}) }}`, "require string keys"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := renderGojaTemplate(t, test.script, test.template)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.message)
		})
	}
}

func TestRenderer_TemplateBridgeUsesFreshDataForEachRender(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)
	value, err := ctx.runtime.RunString(`globalThis.source = {number: 1, read() {return this.number;}}`)
	require.NoError(t, err)
	writeRendererTestTemplate(t, ctx, `{{ data.number }}:{{ data.read() }}`)
	first, err := ctx.RenderTemplate("templates/filter.tpl", value)
	require.NoError(t, err)
	require.Equal(t, "1:1", first)
	_, err = ctx.runtime.RunString(`source.number = 2`)
	require.NoError(t, err)
	second, err := ctx.RenderTemplate("templates/filter.tpl", value)
	require.NoError(t, err)
	require.Equal(t, "2:2", second)
}

func TestRenderer_TemplateBridgeSupportsOpenAPIModels(t *testing.T) {
	for _, version := range []string{"3.0.3", "3.1.0", "3.2.0"} {
		t.Run(version, func(t *testing.T) {
			ctx := newRunnerTemplateTestContext(t)
			spec := fmt.Sprintf(`openapi: %s
info: {title: Users, version: '1'}
servers:
  - url: https://{region}.example.com
    variables:
      region: {default: us, enum: [us, eu]}
paths:
  /users/{id}:
    parameters:
      - {in: path, name: id, required: true, schema: {type: string}}
    get:
      operationId: getUser
      responses:
        '200':
          description: user
          content:
            application/json:
              schema: {$ref: '#/components/schemas/User'}
components:
  schemas:
    User:
      type: object
      properties:
        id: {type: string}
        friend: {$ref: '#/components/schemas/User'}
`, version)
			model, err := ctx.LoadOpenAPIFromString(spec)
			require.NoError(t, err)
			writeRendererTestTemplate(t, ctx, `{{ data.info.title }}|{{ data.servers[0].expandUrl({"region":"eu"}) }}|{% for server in data.servers %}{% for route in server.paths %}{% for operation in route.operations %}{{ operation.method | upper }} {{ route.path }};{% endfor %}{% endfor %}{% endfor %}{% set operation = data.getOperation("GET", "/users/{id}") %}{{ operation.getParameter("path", "id").name }}|{{ operation.pathItem.operations[0].operationId }}|{% set schema = operation.getResponse("200").getMediaType("application/json").schema.reference.target %}{{ schema.getProperty("friend").schema.reference.target.getProperty("id").schema.types[0] }}|{{ data.getOperationById("missing") is none }}|{{ data.raw.info.title }}`)
			rendered, err := ctx.RenderTemplate("templates/filter.tpl", model)
			require.NoError(t, err)
			assert.Equal(t, "Users|https://eu.example.com|GET /users/{id};id|getUser|string|True|Users", rendered)
			raw, err := ctx.ParseOpenAPIFromString(spec)
			require.NoError(t, err)
			writeRendererTestTemplate(t, ctx, `{{ data.info.title }}|{% for path,item in data.paths %}{{ path }}:{{ item["get"].operationId }};{% endfor %}`)
			rendered, err = ctx.RenderTemplate("templates/filter.tpl", raw)
			require.NoError(t, err)
			assert.Equal(t, "Users|/users/{id}:getUser;", rendered)
			writeRendererTestTemplate(t, ctx, `{{ data.servers[0].expandUrl({"region":"invalid"}) }}`)
			_, err = ctx.RenderTemplate("templates/filter.tpl", model)
			require.ErrorContains(t, err, "invalid value")
		})
	}
}

func TestRunner_TemplateBridgeAcrossOutputFunctions(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_TemplateBridgeAcrossOutputFunctions")
	templateDir := filepath.Join(projectDir, "templates")
	require.NoError(t, os.MkdirAll(templateDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "api.tpl"), []byte(`{{ data.info.title }}:{{ data.servers[0].expandUrl() }}:{{ data.getOperation("get", "/users").method }}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "model.proc.ts"), []byte(`import type { Context } from "@ccode/context";
export default function main(ctx: Context) {
 const api = ctx.loadOpenAPIFromString('{"openapi":"3.1.0","info":{"title":"Users","version":"1"},"paths":{"/users":{"get":{"responses":{}}}}}');
 ctx.println(ctx.renderTemplate("templates/api.tpl",api));
 ctx.generate("templates/api.tpl","api.txt",api);
 ctx.accelerate("handler.txt","templates/api.tpl",api);
}`), 0644))
	var output bytes.Buffer
	ctx.stdout = &output
	require.NoError(t, ctx.Run("model"))
	assert.Equal(t, "Users:/:get\n", output.String())
	for _, name := range []string{"api.txt", "handler.txt"} {
		content, err := os.ReadFile(filepath.Join(ctx.config.OutputPath, name))
		require.NoError(t, err)
		assert.Equal(t, "Users:/:get", string(content))
	}
}

func BenchmarkTemplateBridgeSharedGraph(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("objects=%d", size), func(b *testing.B) {
			runtime := goja.New()
			value, err := runtime.RunString(fmt.Sprintf(`(() => {
   const root = {operations: [], title: "API"};
   root.self = root;
   root.read = function() {return this.title;};
   for (let i=0; i<%d; i++) root.operations.push({parent: root, first: root, method: "get", path: "/"+i});
   return root;
  })()`, size))
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := gojaValueToTemplateData(runtime, value); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
