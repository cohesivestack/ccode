package ccode

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/dop251/goja"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCohesiveTemplateEnvironment(t *testing.T) {
	require.False(t, gonja.DefaultEnvironment.Filters.Exists("camelCase"))

	environment, err := newCohesiveTemplateEnvironment()
	require.NoError(t, err)
	require.NotSame(t, gonja.DefaultEnvironment.Filters, environment.Filters)

	for _, spec := range templateFilterSpecs {
		assert.True(t, environment.Filters.Exists(spec.name), "custom filter %q", spec.name)
		assert.False(t, gonja.DefaultEnvironment.Filters.Exists(spec.name), "default filter %q", spec.name)
	}

	assert.True(t, environment.Filters.Exists("lower"))
	assert.Same(t, gonja.DefaultEnvironment.ControlStructures, environment.ControlStructures)
	assert.Same(t, gonja.DefaultEnvironment.Tests, environment.Tests)
	assert.Same(t, gonja.DefaultEnvironment.Context, environment.Context)
}

func TestRegisterTemplateFiltersRejectsExistingFilter(t *testing.T) {
	existing := func(_ *exec.Evaluator, _ *exec.Value, _ *exec.VarArgs) *exec.Value {
		return exec.AsValue("existing")
	}
	filters := exec.NewFilterSet(map[string]exec.FilterFunction{
		"camelCase": existing,
	})

	err := registerTemplateFilters(filters)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `filter "camelCase"`)

	registered, ok := filters.Get("camelCase")
	require.True(t, ok)
	assert.Equal(t, "existing", registered(nil, exec.AsValue("ignored"), exec.NewVarArgs()).String())
}

func TestInitialismTemplateFilterPassesNilOnlyWhenOmitted(t *testing.T) {
	filter := initialismTemplateFilter("probe", func(_ string, initialisms []string) string {
		if initialisms == nil {
			return "nil"
		}
		return fmt.Sprintf("%v", initialisms)
	})

	assert.Equal(t, "nil", filter(nil, exec.AsValue("value"), exec.NewVarArgs()).String())

	params := exec.NewVarArgs()
	params.KwArgs["initialisms"] = exec.AsValue([]string{})
	assert.Equal(t, "[]", filter(nil, exec.AsValue("value"), params).String())

	params = exec.NewVarArgs()
	params.KwArgs["initialisms"] = exec.AsValue([]any{"HTTP", "ID"})
	assert.Equal(t, "[HTTP ID]", filter(nil, exec.AsValue("value"), params).String())
}

func TestRendererStandardAndCustomFilters(t *testing.T) {
	rendered, err := renderTemplateSource(t, `{{ data.name | lower }}|{{ data.name | camelCase }}`, map[string]any{
		"name": "HTTP SERVER",
	})
	require.NoError(t, err)
	assert.Equal(t, "http server|httpServer", rendered)
}

func TestRendererGenericStringFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		input    string
		expected string
	}{
		{name: "camel case", filter: "camelCase", input: "HTTPServer_config", expected: "httpServerConfig"},
		{name: "pascal case", filter: "pascalCase", input: "HTTPServer_config", expected: "HttpServerConfig"},
		{name: "snake case", filter: "snakeCase", input: "HTTPServer_config", expected: "http_server_config"},
		{name: "kebab case", filter: "kebabCase", input: "HTTPServer_config", expected: "http-server-config"},
		{name: "constant case", filter: "constantCase", input: "HTTPServer_config", expected: "HTTP_SERVER_CONFIG"},
		{name: "dot case", filter: "dotCase", input: "HTTPServer_config", expected: "http.server.config"},
		{name: "path case", filter: "pathCase", input: "HTTPServer_config", expected: "http/server/config"},
		{name: "title case", filter: "titleCase", input: "HTTPServer_config", expected: "Http Server Config"},
		{name: "sentence case", filter: "sentenceCase", input: "HTTPServer_config", expected: "Http server config"},
		{name: "upper first", filter: "upperFirst", input: "éclairAccount", expected: "ÉclairAccount"},
		{name: "lower first", filter: "lowerFirst", input: "ÉclairAccount", expected: "éclairAccount"},
		{name: "normalize space", filter: "normalizeSpace", input: "  user\t account\nname  ", expected: "user account name"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := fmt.Sprintf("{{ data.name | %s }}", test.filter)
			rendered, err := renderTemplateSource(t, source, map[string]any{"name": test.input})
			require.NoError(t, err)
			assert.Equal(t, test.expected, rendered)
		})
	}
}

func TestRendererGenericInitialisms(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		input    string
		argument string
		data     map[string]any
		expected string
	}{
		{name: "camel neutral", filter: "camelCase", input: "HTTP server ID", expected: "httpServerId"},
		{name: "pascal neutral", filter: "pascalCase", input: "HTTP server ID", expected: "HttpServerId"},
		{name: "title neutral", filter: "titleCase", input: "HTTP server ID", expected: "Http Server Id"},
		{name: "sentence neutral", filter: "sentenceCase", input: "HTTP server ID", expected: "Http server id"},
		{name: "camel empty", filter: "camelCase", input: "HTTP server ID", argument: "(initialisms=[])", expected: "httpServerId"},
		{name: "camel configured", filter: "camelCase", input: "HTTP server ID", argument: `(initialisms=["HTTP", "ID"])`, expected: "httpServerID"},
		{name: "pascal configured", filter: "pascalCase", input: "HTTP server ID", argument: `(initialisms=["HTTP", "ID"])`, expected: "HTTPServerID"},
		{name: "title configured", filter: "titleCase", input: "HTTP server ID", argument: `(initialisms=["HTTP", "ID"])`, expected: "HTTP Server ID"},
		{name: "sentence configured", filter: "sentenceCase", input: "HTTP server ID", argument: `(initialisms=["HTTP", "ID"])`, expected: "HTTP server ID"},
		{name: "mixed case", filter: "pascalCase", input: "oauth graph ql open api", argument: `(initialisms=["OAuth", "GraphQL", "OpenAPI"])`, expected: "OAuthGraphQLOpenAPI"},
		{name: "data list", filter: "camelCase", input: "HTTP server ID", argument: "(initialisms=data.initialisms)", data: map[string]any{"initialisms": []any{"HTTP", "ID"}}, expected: "httpServerID"},
		{name: "last duplicate casing wins", filter: "pascalCase", input: "http id", argument: `(initialisms=["HTTP", "Id", "id", "iD"])`, expected: "HTTPiD"},
		{name: "complete word only", filter: "camelCase", input: "valid identifier", argument: `(initialisms=["ID"])`, expected: "validIdentifier"},
		{name: "no substring match", filter: "titleCase", input: "identity provider", argument: `(initialisms=["ID"])`, expected: "Identity Provider"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := map[string]any{"name": test.input}
			for key, value := range test.data {
				data[key] = value
			}
			source := fmt.Sprintf("{{ data.name | %s%s }}", test.filter, test.argument)
			rendered, err := renderTemplateSource(t, source, data)
			require.NoError(t, err)
			assert.Equal(t, test.expected, rendered)
		})
	}
}

func TestRendererGoFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		input    string
		argument string
		expected string
	}{
		{name: "exported default", filter: "goExported", input: "user id", expected: "UserID"},
		{name: "unexported default", filter: "goUnexported", input: "get user ID", expected: "getUserID"},
		{name: "exported custom", filter: "goExported", input: "database db", argument: `(initialisms=["DB"])`, expected: "DatabaseDB"},
		{name: "exported custom addition preserves defaults", filter: "goExported", input: "user id database db", argument: `(initialisms=["DB"])`, expected: "UserIDDatabaseDB"},
		{name: "unexported custom", filter: "goUnexported", input: "database db", argument: `(initialisms=["DB"])`, expected: "databaseDB"},
		{name: "mixed case custom", filter: "goExported", input: "oauth client", argument: `(initialisms=["OAuth"])`, expected: "OAuthClient"},
		{name: "default casing override", filter: "goExported", input: "user ids", argument: `(initialisms=["Id"])`, expected: "UserIds"},
		{name: "unexported default casing override", filter: "goUnexported", input: "user id", argument: `(initialisms=["Id"])`, expected: "userId"},
		{name: "plural initialism", filter: "goExported", input: "database dbs", argument: `(initialisms=["DB"])`, expected: "DatabaseDBs"},
		{name: "direct match precedence", filter: "goExported", input: "database dbs", argument: `(initialisms=["DBS"])`, expected: "DatabaseDBS"},
		{name: "both direct and singular", filter: "goExported", input: "database dbs", argument: `(initialisms=["DB", "DBS"])`, expected: "DatabaseDBS"},
		{name: "single suffix removal", filter: "goExported", input: "database dbss", argument: `(initialisms=["DB"])`, expected: "DatabaseDbss"},
		{name: "camel plural", filter: "goExported", input: "dbsGet", argument: `(initialisms=["DB"])`, expected: "DBsGet"},
		{name: "camel nonrecursive plural", filter: "goExported", input: "dbssGet", argument: `(initialisms=["DB"])`, expected: "DbssGet"},
		{name: "unexported plural", filter: "goUnexported", input: "database dbs", argument: `(initialisms=["DB"])`, expected: "databaseDBs"},
		{name: "exported digit", filter: "goExported", input: "123 users", expected: "X123Users"},
		{name: "unexported digit", filter: "goUnexported", input: "123 users", expected: "x123Users"},
		{name: "unexported keyword", filter: "goUnexported", input: "type", expected: "type_"},
		{name: "package", filter: "goPackage", input: "HTTP server", expected: "httpserver"},
		{name: "package keyword", filter: "goPackage", input: "type", expected: "typepkg"},
		{name: "package digit", filter: "goPackage", input: "123 API", expected: "pkg123api"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := fmt.Sprintf("{{ data.name | %s%s }}", test.filter, test.argument)
			rendered, err := renderTemplateSource(t, source, map[string]any{"name": test.input})
			require.NoError(t, err)
			assert.Equal(t, test.expected, rendered)
		})
	}
}

func TestRendererOpenAPIPathFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		input    string
		argument string
		expected string
	}{
		{name: "colon default", filter: "openAPIPathToColon", input: "/users/{userId}", expected: "/users/:userId"},
		{name: "colon omit", filter: "openAPIPathToColon", input: "/users/{userId}", argument: "(omitLeadingSlash=true)", expected: "users/:userId"},
		{name: "colon preserve explicitly", filter: "openAPIPathToColon", input: "/users/{userId}", argument: "(omitLeadingSlash=false)", expected: "/users/:userId"},
		{name: "square brackets", filter: "openAPIPathToSquareBrackets", input: "/users/{userId}", expected: "/users/[userId]"},
		{name: "angle brackets", filter: "openAPIPathToAngleBrackets", input: "/users/{userId}", expected: "/users/<userId>"},
		{name: "dollar", filter: "openAPIPathToDollar", input: "/users/{userId}", expected: "/users/$userId"},
		{name: "static path", filter: "openAPIPathToSquareBrackets", input: "/users/current", argument: "(omitLeadingSlash=true)", expected: "users/current"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := fmt.Sprintf("{{ data.path | %s%s }}", test.filter, test.argument)
			rendered, err := renderTemplateSource(t, source, map[string]any{"path": test.input})
			require.NoError(t, err)
			assert.Equal(t, test.expected, rendered)
		})
	}
}

func TestRendererOpenAPIPathFilterArgumentValidation(t *testing.T) {
	filters := []string{
		"openAPIPathToColon",
		"openAPIPathToSquareBrackets",
		"openAPIPathToAngleBrackets",
		"openAPIPathToDollar",
	}
	tests := []struct {
		name     string
		source   func(string) string
		contains string
	}{
		{name: "non-string input", source: func(filter string) string { return fmt.Sprintf("{{ 42 | %s }}", filter) }, contains: "input must be a string"},
		{name: "positional argument", source: func(filter string) string { return fmt.Sprintf(`{{ data.path | %s(true) }}`, filter) }, contains: "positional arguments are not supported"},
		{name: "unknown keyword", source: func(filter string) string { return fmt.Sprintf(`{{ data.path | %s(strip=true) }}`, filter) }, contains: `unexpected argument "strip"`},
		{name: "non-boolean option", source: func(filter string) string {
			return fmt.Sprintf(`{{ data.path | %s(omitLeadingSlash="true") }}`, filter)
		}, contains: `"omitLeadingSlash" must be a boolean`},
	}

	for _, filter := range filters {
		for _, test := range tests {
			t.Run(filter+" "+test.name, func(t *testing.T) {
				_, err := renderTemplateSource(t, test.source(filter), map[string]any{"path": "/users/{userId}"})
				require.Error(t, err)
				assert.Contains(t, err.Error(), fmt.Sprintf(`filter %q`, filter))
				assert.Contains(t, err.Error(), test.contains)
			})
		}
	}
}

func TestRendererFilterArgumentValidation(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		data      map[string]any
		contains  string
		contains2 string
	}{
		{name: "positional initialisms", source: `{{ data.name | camelCase("HTTP", "ID") }}`, contains: `filter "camelCase"`, contains2: "positional arguments are not supported"},
		{name: "unknown keyword", source: `{{ data.name | camelCase(acronyms=["HTTP"]) }}`, contains: `filter "camelCase"`, contains2: `unexpected argument "acronyms"`},
		{name: "scalar initialisms", source: `{{ data.name | camelCase(initialisms="HTTP") }}`, contains: `filter "camelCase"`, contains2: `"initialisms" must be a list of strings`},
		{name: "numeric initialism", source: `{{ data.name | camelCase(initialisms=[42, "ID"]) }}`, contains: `filter "camelCase"`, contains2: "initialisms[0] must be a string"},
		{name: "null initialism", source: `{{ data.name | camelCase(initialisms=[none, "ID"]) }}`, contains: `filter "camelCase"`, contains2: "initialisms[0] must be a string"},
		{name: "blank initialism", source: `{{ data.name | camelCase(initialisms=["   "]) }}`, contains: `filter "camelCase"`, contains2: "initialisms[0] must not be blank"},
		{name: "argument to no-argument filter", source: `{{ data.name | snakeCase(initialisms=["ID"]) }}`, contains: `filter "snakeCase"`, contains2: `unexpected argument "initialisms"`},
		{name: "positional argument to no-argument filter", source: `{{ data.name | snakeCase("ID") }}`, contains: `filter "snakeCase"`, contains2: "positional arguments are not supported"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := map[string]any{"name": "HTTP server ID"}
			for key, value := range test.data {
				data[key] = value
			}
			_, err := renderTemplateSource(t, test.source, data)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `render template "templates/filter.tpl"`)
			assert.Contains(t, err.Error(), test.contains)
			assert.Contains(t, err.Error(), test.contains2)
		})
	}
}

func TestRendererFilterRejectsNonStringInput(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "number", source: `{{ 42 | camelCase }}`},
		{name: "null", source: `{{ none | camelCase }}`},
		{name: "list", source: `{{ ["HTTP"] | camelCase }}`},
		{name: "object", source: `{{ {"name": "HTTP"} | camelCase }}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := renderTemplateSource(t, test.source, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `filter "camelCase": input must be a string`)
		})
	}
}

func TestRendererGojaInitialismSequence(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)
	writeRendererTestTemplate(t, ctx, `{{ data.name | camelCase(initialisms=data.initialisms) }}`)

	value, err := ctx.runtime.RunString(`({ name: "HTTP server ID", initialisms: ["HTTP", "ID"] })`)
	require.NoError(t, err)
	rendered, err := ctx.RenderTemplate("templates/filter.tpl", value)
	require.NoError(t, err)
	assert.Equal(t, "httpServerID", rendered)

	value, err = ctx.runtime.RunString(`({ name: "HTTP server ID", initialisms: [undefined] })`)
	require.NoError(t, err)
	_, err = ctx.RenderTemplate("templates/filter.tpl", value)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initialisms[0] must be a string")
}

func TestRendererEnvironmentIsReusedWithoutRegistration(t *testing.T) {
	first, err := cohesiveTemplateEnvironment()
	require.NoError(t, err)
	second, err := cohesiveTemplateEnvironment()
	require.NoError(t, err)
	assert.Same(t, first, second)

	ctx := newRunnerTemplateTestContext(t)
	writeRendererTestTemplate(t, ctx, `{{ data.name | camelCase }}`)
	for range 3 {
		rendered, renderErr := ctx.renderTemplate("templates/filter.tpl", map[string]any{"name": "HTTP server"})
		require.NoError(t, renderErr)
		assert.Equal(t, "httpServer", rendered)
	}
}

func TestRendererParallelRenders(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)
	writeRendererTestTemplate(t, ctx, `{{ data.name | goExported(initialisms=["DB"]) }}`)

	const renders = 32
	errors := make(chan error, renders)
	var wait sync.WaitGroup
	for range renders {
		wait.Add(1)
		go func() {
			defer wait.Done()
			rendered, err := ctx.renderTemplate("templates/filter.tpl", map[string]any{"name": "database dbs"})
			if err == nil && rendered != "DatabaseDBs" {
				err = fmt.Errorf("unexpected rendered result %q", rendered)
			}
			errors <- err
		}()
	}
	wait.Wait()
	close(errors)

	for err := range errors {
		assert.NoError(t, err)
	}
}

func renderTemplateSource(t *testing.T, source string, data any) (string, error) {
	t.Helper()

	ctx := newRunnerTemplateTestContext(t)
	writeRendererTestTemplate(t, ctx, source)
	return ctx.renderTemplate("templates/filter.tpl", data)
}

func writeRendererTestTemplate(t *testing.T, ctx *RunnerContext, source string) {
	t.Helper()

	templateDir := filepath.Join(ctx.ccodeContext.config.CCodePath, "templates")
	require.NoError(t, os.MkdirAll(templateDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "filter.tpl"), []byte(source), 0644))
}

func TestRendererGojaUndefinedIsConvertedToNil(t *testing.T) {
	runtime := goja.New()
	value, err := runtime.RunString(`[undefined]`)
	require.NoError(t, err)

	converted, err := gojaValueToTemplateData(value)
	require.NoError(t, err)
	assert.Equal(t, []any{nil}, converted)
}
