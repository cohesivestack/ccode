package ccode

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dop251/goja"
)

func (b *apiModelBuilder) buildModel() *goja.Object {
	root := b.root
	model := b.element("model", root)
	b.set(model, "openapiVersion", apiString(b.child(root, "openapi")))
	info := b.child(root, "info")
	if !info.exists() {
		b.fail("OpenAPI info is required")
	}
	b.set(model, "info", b.info(info))
	components := b.child(root, "components")
	for _, entry := range b.entries(b.child(components, "securitySchemes")) {
		b.securitySchemes[entry.name] = b.securityScheme(entry.loc, entry.name)
	}
	b.rootSecurity = b.security(b.child(root, "security"))
	b.defaultServers = b.serverList(b.child(root, "servers"), "document", true, "document")
	if len(b.defaultServers) == 0 {
		b.defaultServers = []*goja.Object{b.defaultServer()}
	}
	var paths []*goja.Object
	for _, entry := range b.entries(b.child(root, "paths")) {
		if strings.HasPrefix(entry.name, "x-") {
			continue
		}
		if !strings.HasPrefix(entry.name, "/") {
			b.fail("invalid route template %q", entry.name)
			continue
		}
		paths = append(paths, b.path(entry.loc, entry.name, "path", nil, entry.name))
	}
	var webhooks []*goja.Object
	if b.version != "3.0" {
		for _, entry := range b.entries(b.child(root, "webhooks")) {
			if !strings.HasPrefix(entry.name, "x-") {
				webhooks = append(webhooks, b.path(entry.loc, entry.name, "webhook", nil, "webhook:"+entry.name))
			}
		}
	}
	schemas := b.namedSchemas(b.child(components, "schemas"))
	b.set(model, "paths", b.array(paths))
	b.lookup(model, "getPath", paths, "path", false)
	b.set(model, "operations", b.array(b.operations))
	b.operationLookups(model, b.operations)
	b.set(model, "webhooks", b.array(webhooks))
	b.lookup(model, "getWebhook", webhooks, "name", false)
	b.set(model, "schemas", b.array(schemas))
	b.lookup(model, "getSchema", schemas, "name", false)
	b.finalizeServers()
	b.set(model, "servers", b.array(b.servers))
	return model
}

func (b *apiModelBuilder) info(l apiLocation) *goja.Object {
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement("info", v)
	b.copy(o, v, "title", "version", "description", "termsOfService")
	if b.version != "3.0" {
		b.copy(o, v, "summary")
	}
	if c := v.get("contact"); c.exists() {
		b.set(o, "contact", b.simple(c, "contact", "name", "url", "email"))
	}
	if c := v.get("license"); c.exists() {
		names := []string{"name", "url"}
		if b.version != "3.0" {
			names = append(names, "identifier")
		}
		b.set(o, "license", b.simple(c, "license", names...))
	}
	return o
}

func (b *apiModelBuilder) simple(l apiLocation, role string, names ...string) *goja.Object {
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement(role, v)
	b.copy(o, v, names...)
	return o
}

func (b *apiModelBuilder) externalDocs(l apiLocation) *goja.Object {
	return b.simple(l, "externalDocs", "url", "description")
}

func (b *apiModelBuilder) serverList(l apiLocation, scope string, register bool, context string) []*goja.Object {
	var result []*goja.Object
	for _, item := range b.sequence(l) {
		result = append(result, b.server(item, scope, register, context))
	}
	return result
}

var apiServerVariablePattern = regexp.MustCompile(`\{([^{}]+)\}`)

func (b *apiModelBuilder) server(l apiLocation, scope string, register bool, context string) *goja.Object {
	key := "server:" + context + ":" + l.key()
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement(key, v)
	b.cache[key] = o
	b.copy(o, v, "url", "description")
	if b.version == "3.2" {
		b.copy(o, v, "name")
	}
	b.set(o, "scope", scope)
	variables := []*goja.Object{}
	defaults := map[string]string{}
	allowed := map[string]map[string]bool{}
	for _, entry := range b.entries(v.get("variables")) {
		variable := b.element("serverVariable:"+context, entry.loc)
		b.set(variable, "name", entry.name)
		def := b.child(entry.loc, "default")
		if !def.exists() || def.node.Tag != "!!str" {
			b.fail("server variable %q requires a string default at %s", entry.name, entry.loc.key())
		}
		defaultValue := apiString(def)
		b.set(variable, "defaultValue", defaultValue)
		defaults[entry.name] = defaultValue
		if enum := b.child(entry.loc, "enum"); enum.exists() {
			b.set(variable, "enum", b.raw(enum))
			allowed[entry.name] = map[string]bool{}
			for _, item := range b.sequence(enum) {
				allowed[entry.name][apiString(item)] = true
			}
			if !allowed[entry.name][defaultValue] {
				b.fail("server variable %q default is outside its enum", entry.name)
			}
		}
		if desc := b.child(entry.loc, "description"); desc.exists() {
			b.set(variable, "description", b.raw(desc))
		}
		variables = append(variables, variable)
	}
	b.set(o, "variables", b.array(variables))
	b.lookup(o, "getVariable", variables, "name", false)
	template := apiString(v.get("url"))
	if template == "" {
		b.fail("server URL is required at %s", l.key())
	}
	for _, match := range apiServerVariablePattern.FindAllStringSubmatch(template, -1) {
		if _, ok := defaults[match[1]]; !ok {
			b.fail("server URL variable %q has no declaration at %s", match[1], l.key())
		}
	}
	b.set(o, "expandUrl", func(call goja.FunctionCall) goja.Value {
		values := map[string]string{}
		for name, value := range defaults {
			values[name] = value
		}
		argument := call.Argument(0)
		if !goja.IsUndefined(argument) {
			input, ok := argument.(*goja.Object)
			if !ok || input.ClassName() != "Object" {
				panic(b.ctx.runtime.NewGoError(fmt.Errorf("server variables must be an object")))
			}
			for _, name := range input.Keys() {
				if _, ok := defaults[name]; !ok {
					panic(b.ctx.runtime.NewGoError(fmt.Errorf("unknown server variable %q", name)))
				}
				value, ok := input.Get(name).Export().(string)
				if !ok {
					panic(b.ctx.runtime.NewGoError(fmt.Errorf("server variable %q must be a string", name)))
				}
				if choices, exists := allowed[name]; exists && !choices[value] {
					panic(b.ctx.runtime.NewGoError(fmt.Errorf("invalid value %q for server variable %q", value, name)))
				}
				values[name] = value
			}
		}
		return b.ctx.runtime.ToValue(apiServerVariablePattern.ReplaceAllStringFunc(template, func(token string) string { return values[token[1:len(token)-1]] }))
	})
	// Declarations on webhooks/callbacks/links are intentionally excluded from
	// the ordinary-route server catalogue, but still expose empty navigation.
	b.set(o, "paths", b.array(nil))
	b.set(o, "operations", b.array(nil))
	b.lookup(o, "getPath", nil, "path", false)
	b.operationLookups(o, nil)
	if register {
		b.servers = append(b.servers, o)
	}
	return o
}

func (b *apiModelBuilder) defaultServer() *goja.Object {
	if b.implicitServer != nil {
		return b.implicitServer
	}
	// Build a synthetic declaration without pretending it came from the source.
	o := b.object()
	b.set(o, "id", "server:implicit")
	b.set(o, "extensions", b.object())
	b.set(o, "url", "/")
	b.set(o, "scope", "implicit")
	b.set(o, "variables", b.array(nil))
	b.lookup(o, "getVariable", nil, "name", false)
	b.set(o, "expandUrl", func(call goja.FunctionCall) goja.Value {
		arg := call.Argument(0)
		if !goja.IsUndefined(arg) {
			obj, ok := arg.(*goja.Object)
			if !ok || obj.ClassName() != "Object" || len(obj.Keys()) != 0 {
				panic(b.ctx.runtime.NewGoError(fmt.Errorf("implicit server has no variables")))
			}
		}
		return b.ctx.runtime.ToValue("/")
	})
	b.implicitServer = o
	b.servers = append(b.servers, o)
	return o
}

func (b *apiModelBuilder) effectiveServers(declared, inherited []*goja.Object, hasDeclaration bool) []*goja.Object {
	if !hasDeclaration {
		return inherited
	}
	if len(declared) == 0 {
		return []*goja.Object{b.defaultServer()}
	}
	return declared
}

func (b *apiModelBuilder) finalizeServers() {
	for _, server := range b.servers {
		var operations, bindings []*goja.Object
		byPath := map[*goja.Object][]*goja.Object{}
		var pathOrder []*goja.Object
		for _, operation := range b.operations {
			applicable := operation.Get("servers").ToObject(b.ctx.runtime)
			found := false
			for i := 0; i < int(applicable.Get("length").ToInteger()); i++ {
				if applicable.Get(fmt.Sprint(i)) == server {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			operations = append(operations, operation)
			path := operation.Get("pathItem").ToObject(b.ctx.runtime)
			if _, ok := byPath[path]; !ok {
				pathOrder = append(pathOrder, path)
			}
			byPath[path] = append(byPath[path], operation)
		}
		for _, path := range pathOrder {
			binding := b.object()
			b.set(binding, "server", server)
			b.set(binding, "pathItem", path)
			b.set(binding, "path", path.Get("path"))
			b.set(binding, "operations", b.array(byPath[path]))
			b.methodLookup(binding, byPath[path])
			bindings = append(bindings, binding)
		}
		b.set(server, "operations", b.array(operations))
		b.set(server, "paths", b.array(bindings))
		b.lookup(server, "getPath", bindings, "path", false)
		b.operationLookups(server, operations)
	}
}

func (b *apiModelBuilder) path(l apiLocation, name, kind string, parent *goja.Object, context string) *goja.Object {
	key := kind + ":" + context + ":" + name + ":" + l.key()
	if kind == "callbackPath" {
		key += ":" + apiObjectString(parent, "id")
	}
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, true, map[string]bool{})
	o := b.viewElement(key, v)
	b.cache[key] = o
	switch kind {
	case "path":
		b.set(o, "path", name)
	case "webhook":
		b.set(o, "name", name)
	case "callbackPath":
		b.set(o, "expression", name)
		b.set(o, "callback", parent)
	}
	b.copy(o, v, "summary", "description")
	parameters := b.parameters(v.get("parameters"))
	b.set(o, "parameters", b.array(parameters))
	declared := b.serverList(v.get("servers"), "path", kind == "path", key)
	if v.get("servers").exists() {
		b.set(o, "declaredServers", b.array(declared))
	}
	effective := b.effectiveServers(declared, b.defaultServers, v.get("servers").exists())
	// Outbound destinations are defined by webhook registration or callback
	// expressions, not by the ordinary server inheritance chain.
	if kind != "path" {
		effective = nil
	}
	var operations []*goja.Object
	for _, field := range v.order {
		if method, ok := apiStandardMethod(field, b.version); ok && method == field {
			operations = append(operations, b.operation(v.get(field), method, name, kind, o, parameters, effective, apiOperationContext(context, kind, method)))
		} else if field == "additionalOperations" && b.version == "3.2" {
			for _, entry := range b.entries(v.get(field)) {
				if _, standard := apiStandardMethod(entry.name, b.version); standard {
					b.fail("additionalOperations cannot redefine standard method %q", entry.name)
					continue
				}
				operations = append(operations, b.operation(entry.loc, entry.name, name, kind, o, parameters, effective, apiOperationContext(context, kind, entry.name)))
			}
		}
	}
	b.set(o, "operations", b.array(operations))
	b.methodLookup(o, operations)
	return o
}

func (b *apiModelBuilder) operation(l apiLocation, method, path, kind string, parent *goja.Object, pathParams, inheritedServers []*goja.Object, context string) *goja.Object {
	key := "operation:" + kind + ":" + context + ":" + l.key() + ":" + apiObjectString(parent, "id")
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement(key, v)
	b.cache[key] = o
	b.set(o, "method", method)
	b.copy(o, v, "operationId", "summary", "description")
	b.defaultBool(o, v, "deprecated")
	if tags := v.get("tags"); tags.exists() {
		b.set(o, "tags", b.raw(tags))
	} else {
		b.set(o, "tags", b.array(nil))
	}
	if docs := v.get("externalDocs"); docs.exists() {
		b.set(o, "externalDocs", b.externalDocs(docs))
	}
	if id := apiString(v.get("operationId")); id != "" {
		if previous := b.operationIDs[id]; previous != nil && previous != o {
			source := previous.Get("source").ToObject(b.ctx.runtime)
			previousLocation := apiObjectString(source, "documentUri") + "#" + apiObjectString(source, "pointer")
			// Shared callback definitions may have several ownership views of the
			// same declaration. Ordinary routes still require distinct IDs.
			if kind == "path" || previousLocation != l.key() {
				b.fail("duplicate operationId %q", id)
			}
		}
		b.operationIDs[id] = o
	}
	declared := b.parameters(v.get("parameters"))
	b.set(o, "declaredParameters", b.array(declared))
	effective := append([]*goja.Object{}, pathParams...)
	indexes := map[string]int{}
	for i, param := range effective {
		indexes[apiObjectString(param, "location")+"\x00"+apiObjectString(param, "name")] = i
	}
	for _, param := range declared {
		key := apiObjectString(param, "location") + "\x00" + apiObjectString(param, "name")
		if i, ok := indexes[key]; ok {
			effective[i] = param
		} else {
			indexes[key] = len(effective)
			effective = append(effective, param)
		}
	}
	b.set(o, "parameters", b.array(effective))
	b.set(o, "getParameter", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String() + "\x00" + call.Argument(1).String()
		if i, ok := indexes[key]; ok {
			return effective[i]
		}
		return goja.Undefined()
	})
	if body := v.get("requestBody"); body.exists() {
		b.set(o, "requestBody", b.requestBody(body))
	}
	var responses []*goja.Object
	for _, entry := range b.entries(v.get("responses")) {
		if !strings.HasPrefix(entry.name, "x-") {
			responses = append(responses, b.response(entry.loc, entry.name))
		}
	}
	b.set(o, "responses", b.array(responses))
	b.lookup(o, "getResponse", responses, "status", false)
	declaredServers := b.serverList(v.get("servers"), "operation", kind == "path", key)
	if v.get("servers").exists() {
		b.set(o, "declaredServers", b.array(declaredServers))
	}
	security := b.rootSecurity
	if v.get("security").exists() {
		security = b.security(v.get("security"))
		b.set(o, "declaredSecurity", b.array(security))
	}
	b.set(o, "security", b.array(security))
	switch kind {
	case "path":
		b.set(o, "path", path)
		b.set(o, "pathItem", parent)
		b.set(o, "servers", b.array(b.effectiveServers(declaredServers, inheritedServers, v.get("servers").exists())))
		b.operations = append(b.operations, o)
	case "webhook":
		b.set(o, "webhook", parent)
	case "callbackPath":
		b.set(o, "callbackPath", parent)
	}
	var callbacks []*goja.Object
	for _, entry := range b.entries(v.get("callbacks")) {
		callbacks = append(callbacks, b.callback(entry.loc, entry.name, o, context))
	}
	b.set(o, "callbacks", b.array(callbacks))
	b.lookup(o, "getCallback", callbacks, "name", false)
	return o
}

func (b *apiModelBuilder) callback(l apiLocation, name string, owner *goja.Object, context string) *goja.Object {
	// Context is anchored at the ordinary/webhook operation; each usage has
	// its own ownership view and recursive edges stop at an explicit boundary.
	key := "callback:" + context + ":" + name + ":" + l.key() + ":" + apiObjectString(owner, "id")
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement(key, v)
	b.cache[key] = o
	b.set(o, "name", name)
	b.set(o, "owner", owner)
	// A recursive reference cannot be expanded into infinitely many distinct
	// ownership views. Keep an explicit edge to the earlier view at this boundary.
	canonical := l
	seen := map[string]bool{}
	for b.child(canonical, "$ref").exists() && !seen[canonical.key()] {
		seen[canonical.key()] = true
		canonical = b.resolve(canonical)
	}
	activeKey := context + ":" + canonical.key()
	if target := b.activeCallbacks[activeKey]; target != nil {
		b.set(o, "recursiveTarget", target)
		b.set(o, "paths", b.array(nil))
		b.lookup(o, "getPath", nil, "expression", false)
		return o
	}
	b.activeCallbacks[activeKey] = o
	defer delete(b.activeCallbacks, activeKey)
	var paths []*goja.Object
	for _, expr := range v.order {
		if !strings.HasPrefix(expr, "x-") {
			paths = append(paths, b.path(v.get(expr), expr, "callbackPath", o, context))
		}
	}
	b.set(o, "paths", b.array(paths))
	b.lookup(o, "getPath", paths, "expression", false)
	return o
}

func (b *apiModelBuilder) parameters(l apiLocation) []*goja.Object {
	var result []*goja.Object
	seen := map[string]bool{}
	for _, entry := range b.sequence(l) {
		param := b.parameter(entry, "", false)
		key := apiObjectString(param, "location") + "\x00" + apiObjectString(param, "name")
		if seen[key] {
			b.fail("duplicate parameter (%s, %s) at %s", apiObjectString(param, "location"), apiObjectString(param, "name"), entry.key())
		}
		seen[key] = true
		result = append(result, param)
	}
	return result
}

func (b *apiModelBuilder) security(l apiLocation) []*goja.Object {
	var requirements []*goja.Object
	for _, entry := range b.sequence(l) {
		o := b.element("securityRequirement", entry)
		var uses []*goja.Object
		for _, scheme := range b.entries(entry) {
			definition := b.securitySchemes[scheme.name]
			if definition == nil {
				b.fail("unknown security scheme %q at %s", scheme.name, scheme.loc.key())
				continue
			}
			use := b.object()
			b.set(use, "name", scheme.name)
			b.set(use, "scheme", definition)
			b.set(use, "scopes", b.raw(scheme.loc))
			uses = append(uses, use)
		}
		b.set(o, "schemes", b.array(uses))
		requirements = append(requirements, o)
	}
	return requirements
}

func (b *apiModelBuilder) securityScheme(l apiLocation, name string) *goja.Object {
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement("securityScheme:"+name, v)
	b.set(o, "name", name)
	b.copy(o, v, "type", "description")
	if b.version == "3.2" {
		b.copy(o, v, "deprecated")
	}
	switch apiString(v.get("type")) {
	case "apiKey":
		b.set(o, "location", b.raw(v.get("in")))
		b.set(o, "parameterName", b.raw(v.get("name")))
	case "http":
		b.copy(o, v, "scheme", "bearerFormat")
	case "openIdConnect":
		b.copy(o, v, "openIdConnectUrl")
	case "mutualTLS":
		if b.version == "3.0" {
			b.fail("mutualTLS is not supported by OpenAPI 3.0")
		}
	case "oauth2":
		if b.version == "3.2" {
			b.copy(o, v, "oauth2MetadataUrl")
		}
		var flows []*goja.Object
		for _, entry := range b.entries(v.get("flows")) {
			switch entry.name {
			case "implicit", "password", "clientCredentials", "authorizationCode":
			case "deviceAuthorization":
				if b.version != "3.2" {
					b.fail("deviceAuthorization requires OpenAPI 3.2")
				}
			default:
				if !strings.HasPrefix(entry.name, "x-") {
					b.fail("unknown OAuth flow %q", entry.name)
				}
				continue
			}
			flow := b.simple(entry.loc, "oauthFlow", "authorizationUrl", "tokenUrl", "refreshUrl")
			b.set(flow, "kind", entry.name)
			scopes := b.child(entry.loc, "scopes")
			if scopes.exists() {
				b.set(flow, "scopes", b.raw(scopes))
			} else {
				b.set(flow, "scopes", b.object())
			}
			if b.version == "3.2" {
				if device := b.child(entry.loc, "deviceAuthorizationUrl"); device.exists() {
					b.set(flow, "deviceAuthorizationUrl", b.raw(device))
				}
			}
			flows = append(flows, flow)
		}
		b.set(o, "flows", b.array(flows))
		b.lookup(o, "getFlow", flows, "kind", false)
	default:
		b.fail("unsupported security scheme at %s", l.key())
	}
	return o
}

// Nested callbacks retain the context of their originating route operation.
func apiOperationContext(context, kind, method string) string {
	if kind == "callbackPath" {
		return context
	}
	return context + ":" + method
}
