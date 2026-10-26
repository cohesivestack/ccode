package ccode

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func checkOpenAPIModel(t *testing.T, ctx *RunnerContext, model goja.Value, script string) {
	t.Helper()
	require.NoError(t, ctx.runtime.Set("api", model))
	_, err := ctx.runtime.RunString(`function check(value, message) { if (!value) throw new Error(message); }
 ` + script)
	require.NoError(t, err)
}

func TestOpenAPIModel_VersionedNavigation(t *testing.T) {
	for _, version := range []string{"3.0.3", "3.1.0", "3.2.0"} {
		t.Run(version, func(t *testing.T) {
			ctx := newRunnerOpenAPITestContext(t)
			document := fmt.Sprintf(`openapi: %s
info: {title: Users, version: 1.0.0}
servers:
  - url: https://{region}.example.com/v1
    variables:
      region: {default: us, enum: [us, eu]}
  - url: https://backup.example.com/v1
security:
  - Key: []
paths:
  /users/{id}:
    parameters:
      - {$ref: '#/components/parameters/Id'}
      - {name: X-Tenant, in: header, required: true, schema: {type: string}}
    get:
      operationId: getUser
      parameters:
        - {name: X-Tenant, in: header, required: false, schema: {type: string}}
        - {name: expand, in: query, schema: {type: boolean}}
      responses:
        '200': {$ref: '#/components/responses/User'}
        2XX: {description: success}
        default: {description: fallback}
    patch:
      operationId: updateUser
      servers: [{url: https://writes.example.com/v1}]
      security: []
      requestBody:
        required: true
        content:
          application/json: {schema: {$ref: '#/components/schemas/User'}}
      responses: {'204': {description: updated}}
  /health:
    servers: [{url: /internal}]
    get:
      responses: {'200': {description: healthy}}
components:
  parameters:
    Id: {name: id, in: path, required: true, schema: {type: string}}
  responses:
    User:
      description: user
      headers:
        X-Rate: {schema: {type: integer}}
      content:
        application/json: {schema: {$ref: '#/components/schemas/User'}}
  schemas:
    User:
      type: object
      required: [id]
      properties:
        id: {type: string}
        displayName: {type: string}
      additionalProperties: false
      x-code-name: UserDTO
  securitySchemes:
    Key: {type: apiKey, in: header, name: X-API-Key}
`, version)
			model, err := ctx.LoadOpenAPIFromString(document)
			require.NoError(t, err)
			checkOpenAPIModel(t, ctx, model, `
const get = api.getOperation("GET", "/users/{id}");
const patch = api.getOperationById("updateUser");
check(api.paths.length === 2 && api.operations.length === 3, "ordinary collections");
check(get === api.paths[0].getOperation("get"), "shared path operation");
check(api.servers.length === 4, "all scoped server declarations");
check(api.servers[0].operations.length === 1, "root server applicability");
check(api.servers[0].getPath("/users/{id}").operations[0] === get, "scoped path operation");
check(api.servers[0].getPath("/users/{id}").pathItem.operations.length === 2, "original path is shared");
check(api.servers[0].getOperation("PATCH", "/users/{id}") === undefined, "overridden server excluded");
check(api.servers[0].getOperationById("getUser") === get, "server operation id lookup");
check(patch.servers[0].scope === "operation", "operation override");
check(api.getOperation("get", "/health").servers[0].scope === "path", "path override");
check(get.parameters.length === 3 && get.declaredParameters.length === 2, "effective parameters");
check(get.getParameter("path", "id").reference.uri === "#/components/parameters/Id", "resolved parameter provenance");
check(get.getParameter("header", "X-Tenant").required === false, "parameter override");
check(get.pathItem.parameters[1].required === true, "declarations unchanged");
check(get.getParameter("header", "x-tenant") === undefined, "parameter keys are exact");
check(get.security[0].schemes[0].scheme.parameterName === "X-API-Key", "effective security");
check(patch.security.length === 0 && patch.declaredSecurity.length === 0, "security opt out");
check(get.declaredSecurity === undefined, "undeclared security remains absent");
const response = get.getResponse("200");
check(response.getHeader("x-rate").name === "X-Rate", "case-insensitive header lookup");
check(get.getResponse("201") === undefined && get.getResponse("2XX").status === "2XX", "response lookup exact");
check(response.getMediaType("application/json").schema.reference.target === api.getSchema("User").schema, "schema target identity");
const user = api.getSchema("User").schema;
check(user.getProperty("id").required && !user.getProperty("displayName").required, "property required membership");
check(user.getProperty("id").parent === user, "parent identity");
check(user.additionalProperties.kind === "boolean" && user.additionalProperties.value === false, "boolean additionalProperties");
check(user.extensions["x-code-name"] === "UserDTO", "extensions");
check(patch.requestBody.required && patch.requestBody.getMediaType("application/json"), "request body");
check(api.servers[0].expandUrl() === "https://us.example.com/v1", "server defaults");
check(api.servers[0].expandUrl({region: "eu"}) === "https://eu.example.com/v1", "server overrides");
for (const variables of [{region: "bad"}, {unknown: "x"}, null, {region: 1}]) {
 let threw = false; try { api.servers[0].expandUrl(variables); } catch (_) { threw = true; }
 check(threw, "invalid variable input rejected");
}
check(Object.isFrozen(api) && Object.isFrozen(api.operations) && Object.isFrozen(user), "immutable model");
check(Object.isFrozen(api.raw.paths["/users/{id}"].parameters), "deep immutable raw");
check(api.raw.paths["/users/{id}"].get.responses["200"].$ref !== undefined, "raw reference retained");
check(api.raw.paths["/users/{id}"].get.responses["200"].content === undefined, "raw is not materialized");
check(api.getPath("/missing") === undefined && api.getSchema("missing") === undefined, "absent lookups");
check(api.getOperationById("") === undefined, "absent operationId is not indexed");
`)
		})
	}
}

func TestOpenAPIModel_ExternalReferencesAndCycles(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "api.yaml", `openapi: 3.1.0
info: {title: External, version: '1'}
paths:
  /users: {$ref: './paths.yaml#/Users'}
components:
  schemas:
    User: {$ref: './schema.yaml'}
`)
	writeOpenAPITestFile(t, ctx, "paths.yaml", `Users:
  get:
    operationId: users
    responses:
      '200': {$ref: './response.yaml'}
`)
	writeOpenAPITestFile(t, ctx, "response.yaml", `description: users
content:
  application/json:
    schema: {$ref: './schema.yaml'}
    examples:
      sample:
        value: {$ref: './does-not-exist.yaml', $id: 'not-a-schema'}
`)
	writeOpenAPITestFile(t, ctx, "schema.yaml", `type: object
properties:
  friend: {$ref: '#'}
  escaped/name: {type: string}
`)
	model, err := ctx.LoadOpenAPIFromFile("api.yaml")
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
const op = api.getOperationById("users");
const ref = op.getResponse("200").getMediaType("application/json").schema;
const user = api.getSchema("User").schema.reference.target;
check(ref.reference.target === user, "external shared target");
check(user.getProperty("friend").schema.reference.target === user, "recursive target");
check(user.getProperty("escaped/name").source.pointer === "/properties/escaped~1name", "escaped pointer");
check(user.source.documentUri.endsWith("/schema.yaml") && user.source.pointer === "", "target source");
check(api.getSchema("User").schema.reference.source.pointer === "/components/schemas/User/$ref", "reference source");
check(api.getSchema("User").schema.raw.$ref === "./schema.yaml", "original declaration");
check(api.paths[0].reference.targetSource.pointer === "/Users", "path reference provenance");
check(op.source.documentUri.endsWith("/paths.yaml") && op.source.line === 3, "operation source");
check(op.getResponse("200").getMediaType("application/json").getExample("sample").value.$ref === "./does-not-exist.yaml", "example refs are data");
`)
}

func TestOpenAPIModel_SchemaDialectsAndVersionSemantics(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	model, err := ctx.LoadOpenAPIFromString(`openapi: 3.1.0
info: {title: Schemas, version: '1'}
jsonSchemaDialect: https://json-schema.org/draft/2020-12/schema
paths: {}
components:
  schemas:
    Root:
      $id: https://example.com/schema
      $defs:
        String: {$anchor: text, type: string}
      type: object
      properties:
        anchored: {$ref: '#text', maxLength: 4}
        local: {$ref: '#/$defs/String'}
        dynamic: {$dynamicRef: '#node'}
        mixed: {type: [string, 'null'], default: null, enum: []}
        scoped:
          $id: nested
          $defs:
            Number: {type: number, exclusiveMinimum: 0}
          properties:
            n: {$ref: '#/$defs/Number'}
      allOf: [{properties: {composed: {type: string}}}]
    Allow: true
    Deny: false
`)
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
const root = api.getSchema("Root").schema;
const anchored = root.getProperty("anchored").schema;
check(root.dialect === "https://json-schema.org/draft/2020-12/schema", "effective dialect");
check(anchored.reference.target === root.getDefinition("String").schema, "anchor resolution");
check(anchored.maxLength === 4 && anchored.types.length === 0, "ref siblings are separate constraints");
check(root.getProperty("local").schema.reference.target === anchored.reference.target, "resource-relative pointer");
check(root.getProperty("dynamic").schema.dynamicRef === "#node", "dynamic reference retained");
const mixed = root.getProperty("mixed").schema;
check(mixed.types[1] === "null" && mixed.default === null && mixed.enum.length === 0, "null and empty enum preserved");
const scoped = root.getProperty("scoped").schema;
check(scoped.getProperty("n").schema.reference.target === scoped.getDefinition("Number").schema, "nested id scope");
check(scoped.getDefinition("Number").schema.exclusiveMinimum === 0, "numeric exclusive bound");
check(root.getProperty("composed") === undefined && root.allOf[0].getProperty("composed"), "composition not flattened");
check(api.getSchema("Allow").schema.value && !api.getSchema("Deny").schema.value, "boolean schemas");
`)
	model, err = ctx.LoadOpenAPIFromString(`openapi: 3.0.3
info: {title: Legacy, version: '1'}
paths: {}
components:
  schemas:
    Number: {type: number, minimum: 0, exclusiveMinimum: true, nullable: true}
    Reference: {$ref: '#/components/schemas/Number', maximum: 10}
`)
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
const number = api.getSchema("Number").schema;
check(number.nullable === true && number.exclusiveMinimum === true, "3.0 semantics");
check(api.getSchema("Reference").schema.maximum === undefined, "3.0 ref siblings ignored");
check(api.getSchema("Reference").schema.reference.target === number, "3.0 ref target");
`)
}

func TestOpenAPIModel_WebhooksCallbacksAnd32(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	model, err := ctx.LoadOpenAPIFromString(`openapi: 3.2.0
info: {title: Events, version: '1'}
servers: [{url: https://example.com, name: main}]
paths:
  /events:
    query:
      operationId: queryEvents
      parameters: [{name: filter, in: querystring, content: {application/json: {schema: {type: object}}}}]
      responses:
        '200':
          summary: event stream
          description: stream
          content:
            application/jsonl:
              itemSchema: {type: object}
              examples:
                first: {dataValue: {id: 1}, serializedValue: '{"id":1}'}
    post:
      operationId: createEvent
      responses: {'202': {description: accepted}}
      callbacks:
        done:
          '{$request.body#/callbackUrl}':
            post:
              operationId: eventDone
              responses: {'204': {description: acknowledged}}
    additionalOperations:
      REPORT:
        operationId: reportEvents
        responses: {'200': {description: report}}
webhooks:
  changed:
    post:
      operationId: eventChanged
      responses: {'204': {description: acknowledged}}
`)
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
check(api.operations.length === 3 && api.servers[0].operations.length === 3, "ordinary operations only");
check(api.servers[0].name === "main", "3.2 server name");
const query = api.getOperation("QUERY", "/events");
check(query.method === "query" && query.getParameter("querystring", "filter"), "3.2 query");
check(api.getOperation("REPORT", "/events").operationId === "reportEvents", "additional method");
check(api.getOperation("report", "/events") === undefined, "additional method case preserved");
const media = query.getResponse("200").getMediaType("application/jsonl");
check(media.itemSchema.types[0] === "object" && media.getExample("first").dataValue.id === 1, "3.2 media types");
check(query.getResponse("200").summary === "event stream", "3.2 response summary");
const post = api.getOperationById("createEvent");
const callback = post.getCallback("done");
const destination = callback.getPath("{$request.body#/callbackUrl}");
check(callback.owner === post && destination.callback === callback, "callback ownership");
check(destination.getOperation("POST").callbackPath === destination, "callback operation ownership");
check(api.getOperationById("eventDone") === undefined && api.getOperationById("eventChanged") === undefined, "outbound operations not flattened");
check(api.getWebhook("changed").getOperation("post").webhook === api.webhooks[0], "webhook navigation");
`)
}

func TestOpenAPIModel_DefaultServersAndErrors(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	model, err := ctx.LoadOpenAPIFromBytes([]byte(`openapi: 3.1.0
info: {title: Defaults, version: '1'}
paths:
  /x:
    get:
      responses: {'200': {description: ok}}
`))
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
check(api.servers.length === 1 && api.servers[0].scope === "implicit", "implicit server");
check(api.servers[0].raw === undefined && api.servers[0].source === undefined, "synthetic metadata");
check(api.servers[0].expandUrl() === "/" && api.servers[0].operations[0] === api.operations[0], "implicit navigation");
`)
	_, err = ctx.LoadOpenAPIFromString("openapi: 3.1.0\ninfo: {title: X, version: '1'}\npaths: {}", ctx.runtime.ToValue(map[string]any{"expectedVersion": "3.0"}))
	require.ErrorContains(t, err, "expected OpenAPI 3.0.x")
	for _, test := range []struct{ name, body, contains string }{
		{"duplicate IDs", `paths:
  /x:
    get: {operationId: same, responses: {}}
    post: {operationId: same, responses: {}}`, "duplicate operationId"},
		{"duplicate parameters", `paths:
  /x:
    parameters:
      - {in: query, name: q, schema: {type: string}}
      - {in: query, name: q, schema: {type: string}}
    get: {responses: {}}`, "duplicate parameter"},
		{"missing reference", `paths: {}
components:
  schemas:
    X: {$ref: '#/components/schemas/Missing'}`, "resolve reference"},
		{"missing security", `paths: {}
security: [{Missing: []}]`, "unknown security scheme"},
		{"bad bounds", `paths: {}
components:
  schemas:
    X: {type: number, exclusiveMinimum: true}`, "must be numeric"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := ctx.LoadOpenAPIFromString("openapi: 3.1.0\ninfo: {title: X, version: '1'}\n" + test.body)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.contains)
		})
	}
	var nilContext *RunnerContext
	_, err = nilContext.LoadOpenAPIFromString("")
	require.ErrorContains(t, err, "not initialized")
}

func TestRunner_OpenAPIModelNativeLoaders(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunner_OpenAPIModelNativeLoaders")
	spec := `{"openapi":"3.1.0","info":{"title":"Native","version":"1"},"paths":{"/x":{"get":{"responses":{"200":{"description":"ok"}}}}}}`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "api.json"), []byte(spec), 0644))
	script := `import type { Context } from "@ccode/context";
import type * as OpenAPI from "@ccode/openapi";
import * as OpenAPIRaw from "@ccode/openapi/raw";
export default function main(ctx: Context) {
  const api: OpenAPI.V3_1.ApiModel = ctx.loadOpenAPIFromFile("api.json", {expectedVersion: "3.1"});
  const text = SPEC;
  const fromString = ctx.loadOpenAPIFromString(text, {expectedVersion: "3.1"});
  const fromBytes = ctx.loadOpenAPIFromBytes(Array.from(text).map(c => c.charCodeAt(0)));
  const raw: OpenAPIRaw.Document = ctx.parseOpenAPIFromString(text);
  ctx.println(JSON.stringify({method: api.servers[0].getPath("/x")?.getOperation("GET")?.method,
    version: raw.openapi, stringTitle: fromString.info.title, bytesTitle: fromBytes.info.title,
    rawMethod: OpenAPIRaw.V3_0.HttpMethods.GET, frozen: Object.isFrozen(api.raw)}));
}`
	script = strings.Replace(script, "SPEC", fmt.Sprintf("%q", spec), 1)
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "model.proc.ts"), []byte(script), 0644))
	var output bytes.Buffer
	ctx.stdout = &output
	require.NoError(t, ctx.Run("model"))
	require.JSONEq(t, `{"method":"get","version":"3.1.0","stringTitle":"Native","bytesTitle":"Native","rawMethod":"get","frozen":true}`, output.String())
}

func TestOpenAPIModel_MetadataAndNestedEncodings(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	model, err := ctx.LoadOpenAPIFromString(`openapi: 3.2.0
info:
  title: Metadata
  version: '1'
  summary: API summary
  contact: {name: Support, email: support@example.com, url: https://example.com}
  license: {name: MIT, identifier: MIT}
security:
  - OAuth: [read]
    Key: []
  - {}
components:
  securitySchemes:
    Key: {type: apiKey, in: cookie, name: session}
    Basic: {type: http, scheme: basic}
    Mutual: {type: mutualTLS}
    Connect: {type: openIdConnect, openIdConnectUrl: https://example.com/.well-known/openid-configuration}
    OAuth:
      type: oauth2
      deprecated: true
      oauth2MetadataUrl: https://example.com/.well-known/oauth-authorization-server
      flows:
        deviceAuthorization:
          deviceAuthorizationUrl: https://example.com/device
          tokenUrl: https://example.com/token
          scopes: {read: Read access}
paths:
  /upload:
    post:
      externalDocs: {url: https://example.com/docs, description: Help}
      requestBody:
        content:
          multipart/mixed:
            schema: {type: array}
            prefixEncoding:
              - contentType: application/json
                headers: {X-Part: {schema: {type: string}}}
                encoding: {nested: {style: form}}
            itemEncoding: {contentType: application/octet-stream}
          multipart/form-data:
            schema:
              type: object
              discriminator: {propertyName: kind, mapping: {file: '#/components/schemas/File'}, defaultMapping: Other}
              xml: {name: upload, nodeType: element}
            encoding:
              file: {contentType: image/png, explode: true}
      responses:
        '201':
          description: created
          links:
            next:
              operationId: next
              parameters: {id: '$response.body#/id'}
              requestBody: {$ref: literal-payload, schema: {$id: arbitrary-data}}
              server: {url: https://links.example.com}
`)
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
check(api.info.contact.email === "support@example.com" && api.info.license.identifier === "MIT", "info metadata");
const op = api.operations[0];
check(op.externalDocs.description === "Help", "external documentation");
check(op.security.length === 2 && op.security[0].schemes.length === 2 && op.security[1].schemes.length === 0, "security OR and AND");
const oauth = op.security[0].schemes[0].scheme;
check(oauth.deprecated && oauth.getFlow("deviceAuthorization").scopes.read === "Read access", "OAuth flow");
check(oauth.getFlow("deviceAuthorization").deviceAuthorizationUrl.endsWith("/device"), "device authorization URL");
const mixed = op.requestBody.getMediaType("multipart/mixed");
check(mixed.prefixEncoding[0].headers[0].schema.types[0] === "string", "prefix encoding headers");
check(mixed.prefixEncoding[0].encoding[0].encoding.style === "form", "nested encodings");
check(mixed.itemEncoding.contentType === "application/octet-stream", "item encoding");
const form = op.requestBody.getMediaType("multipart/form-data");
check(form.getEncoding("file").contentType === "image/png" && form.getEncoding("missing") === undefined, "named encoding lookup");
check(form.schema.xml.nodeType === "element" && form.schema.discriminator.defaultMapping === "Other", "schema metadata");
const link = op.getResponse("201").getLink("next");
check(link.parameters.id === "$response.body#/id" && link.requestBody.$ref === "literal-payload", "link expressions and data");
check(link.server.expandUrl() === "https://links.example.com" && link.server.operations.length === 0, "link server");
check(api.servers.length === 1 && api.servers[0].scope === "implicit", "link servers excluded from root catalogue");
`)
}

func TestOpenAPIModel_ScopedDeclarationsAndExternalSchemaFragments(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	writeOpenAPITestFile(t, ctx, "api.yaml", `openapi: 3.2.0
$self: https://api.example.com/openapi.yaml
info: {title: Scopes, version: '1'}
servers: [{url: /same}]
paths:
  /root:
    get: {responses: {}}
  /path:
    servers: [{url: /same}]
    get: {responses: {}}
  /empty:
    servers: []
    get: {responses: {}}
components:
  schemas:
    Declared: {type: string}
    Internal: {$ref: '#/components/schemas/Declared'}
`)
	model, err := ctx.LoadOpenAPIFromFile("api.yaml")
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
check(api.servers.length === 3, "same URL declarations stay distinct");
check(api.servers[0] !== api.servers[1] && api.servers[0].url === api.servers[1].url, "scoped identity");
check(api.servers[0].operations[0].path === "/root" && api.servers[1].operations[0].path === "/path", "scoped applicability");
check(api.getPath("/empty").declaredServers.length === 0 && api.servers[2].scope === "implicit", "empty override");
check(api.getSchema("Internal").schema.reference.resolvedUri === "https://api.example.com/openapi.yaml#/components/schemas/Declared", "$self establishes base");
check(api.getSchema("Internal").schema.reference.target === api.getSchema("Declared").schema, "$self resource resolution");
`)
	writeOpenAPITestFile(t, ctx, "fragment-api.yaml", `openapi: 3.1.0
info: {title: Fragments, version: '1'}
paths: {}
components:
  schemas:
    External: {$ref: './bundle.yaml#/Thing'}
`)
	writeOpenAPITestFile(t, ctx, "bundle.yaml", `Thing:
  $id: https://example.com/thing
  $schema: https://json-schema.org/draft/2020-12/schema
  $defs:
    Text: {$anchor: text, type: string}
  properties:
    value: {$ref: '#text'}
`)
	model, err = ctx.LoadOpenAPIFromFile("fragment-api.yaml")
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
const thing = api.getSchema("External").schema.reference.target;
check(thing.dialect === "https://json-schema.org/draft/2020-12/schema", "fragment dialect");
check(thing.getProperty("value").schema.reference.target === thing.getDefinition("Text").schema, "fragment resource anchor");
`)
}

func TestOpenAPIModel_RecursiveCallbacks(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	model, err := ctx.LoadOpenAPIFromString(`openapi: 3.1.0
info: {title: Recursive callbacks, version: '1'}
paths:
  /subscribe:
    post:
      responses: {'202': {description: accepted}}
      callbacks:
        next: {$ref: '#/components/callbacks/Next'}
components:
  callbacks:
    Next:
      '{$request.body#/url}':
        post:
          responses: {'204': {description: ok}}
          callbacks:
            next: {$ref: '#/components/callbacks/Next'}
`)
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
const outer = api.operations[0].getCallback("next");
const nestedOp = outer.paths[0].operations[0];
const nested = nestedOp.getCallback("next");
check(nested.owner === nestedOp, "recursive boundary owner");
check(nested.paths.length === 0 && nested.recursiveTarget === outer, "recursive callback boundary");
`)
}

func TestOpenAPIModel_SharedCallbackDefinitionsHaveDistinctOwnership(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	model, err := ctx.LoadOpenAPIFromString(`openapi: 3.1.0
info: {title: Shared callbacks, version: '1'}
paths:
  /subscribe:
    post:
      responses: {}
      callbacks:
        first: {$ref: '#/components/callbacks/Notify'}
        second: {$ref: '#/components/callbacks/Notify'}
components:
  callbacks:
    Notify:
      '{$request.body#/url}':
        post:
          operationId: notify
          responses: {}
`)
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
const op = api.operations[0];
const first = op.getCallback("first");
const second = op.getCallback("second");
check(first.paths[0] !== second.paths[0], "distinct ownership views");
check(first.paths[0].callback === first && second.paths[0].callback === second, "path ownership");
check(first.paths[0].operations[0].callbackPath === first.paths[0], "first operation owner");
check(second.paths[0].operations[0].callbackPath === second.paths[0], "second operation owner");
`)
}

func TestOpenAPIModel_RawValuesRemainJSONData(t *testing.T) {
	ctx := newRunnerOpenAPITestContext(t)
	model, err := ctx.LoadOpenAPIFromString(`openapi: 3.1.0
info: {title: JSON values, version: '1'}
paths: {}
components:
  schemas:
    Date:
      type: string
      default: 2026-10-19
      examples: [null, true, 12, text]
`)
	require.NoError(t, err)
	checkOpenAPIModel(t, ctx, model, `
const date = api.getSchema("Date").schema;
check(typeof date.default === "string", "YAML timestamp converted to JSON data");
check(date.examples[0] === null && date.examples[1] === true && date.examples[2] === 12, "JSON primitive values");
check(typeof api.raw.components.schemas.Date.default === "string", "raw values contain no Go wrappers");
`)
	_, err = ctx.LoadOpenAPIFromString("openapi: 3.1.0\ninfo: {title: Alias, version: '1'}\npaths: {}\nx-cycle: &cycle {self: *cycle}\n")
	require.ErrorContains(t, err, "cyclic YAML alias")
}
