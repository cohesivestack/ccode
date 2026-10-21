package ccode

// The ergonomic API is built directly into the runner's Goja runtime. Objects
// are allocated before their children, so cycles and shared graph identity do
// not need a lossy JSON round trip. Public contracts live in template/openapi.
import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	"gopkg.in/yaml.v3"
)

type apiLocation struct {
	node                                   *yaml.Node
	schema                                 bool
	documentURI, pointer, baseURI, dialect string
}

func (l apiLocation) exists() bool { return l.node != nil }

func (l apiLocation) key() string { return l.documentURI + "#" + l.pointer }

type apiModelBuilder struct {
	ctx             *RunnerContext
	version         string
	err             error
	root            apiLocation
	documents       map[string]apiLocation
	resources       map[string]apiLocation
	anchors         map[string]apiLocation
	locations       map[*yaml.Node]apiLocation
	rawValues       map[*yaml.Node]goja.Value
	cache           map[string]*goja.Object
	activeCallbacks map[string]*goja.Object
	objects         []*goja.Object
	operations      []*goja.Object
	servers         []*goja.Object
	operationIDs    map[string]*goja.Object
	securitySchemes map[string]*goja.Object
	rootSecurity    []*goja.Object
	defaultServers  []*goja.Object
	implicitServer  *goja.Object
}

func (ctx *RunnerContext) LoadOpenAPIFromFile(filePath string, options ...goja.Value) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}
	path, err := filepath.Abs(filepath.Join(ctx.ccodeContext.config.CCodePath, filePath))
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read OpenAPI file %s: %w", filePath, err)
	}
	return ctx.loadOpenAPIModel(data, (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String(), options)
}

func (ctx *RunnerContext) LoadOpenAPIFromString(spec string, options ...goja.Value) (goja.Value, error) {
	return ctx.loadOpenAPIModel([]byte(spec), "memory:openapi", options)
}

func (ctx *RunnerContext) LoadOpenAPIFromBytes(spec []byte, options ...goja.Value) (goja.Value, error) {
	return ctx.loadOpenAPIModel(spec, "memory:openapi", options)
}

func (ctx *RunnerContext) loadOpenAPIModel(data []byte, uri string, options []goja.Value) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}
	expected, err := ctx.parseOpenAPIFileOptions(options)
	if err != nil {
		return nil, err
	}
	// libopenapi builds the structural v3 model from a reference-free copy.
	// The source graph below resolves references with their physical locations
	// and JSON Schema resource scopes. This also avoids interpreting literal
	// $ref keys inside example payloads as references to load.
	validationBytes, err := apiModelValidationBytes(data, uri)
	if err != nil {
		return nil, err
	}
	config := datamodel.NewDocumentConfiguration()
	config.AllowFileReferences = false
	config.AllowRemoteReferences = false
	document, err := libopenapi.NewDocumentWithConfiguration(validationBytes, config)
	if err != nil {
		return nil, fmt.Errorf("parse OpenAPI document: %w", err)
	}
	defer document.Release()
	info := document.GetSpecInfo()
	if info == nil {
		return nil, fmt.Errorf("OpenAPI document has no specification information")
	}
	version := openAPIMajorMinorVersion(info.Version)
	switch version {
	case "3.0", "3.1", "3.2":
	default:
		return nil, fmt.Errorf("unsupported OpenAPI version %q", info.Version)
	}
	if expected != "" && expected != version {
		return nil, fmt.Errorf("expected OpenAPI %s.x, but %s declares %s", expected, uri, info.Version)
	}
	if _, err = document.BuildV3Model(); err != nil {
		return nil, fmt.Errorf("build OpenAPI model: %w", err)
	}
	b := &apiModelBuilder{
		ctx:             ctx,
		version:         version,
		documents:       map[string]apiLocation{},
		resources:       map[string]apiLocation{},
		anchors:         map[string]apiLocation{},
		locations:       map[*yaml.Node]apiLocation{},
		rawValues:       map[*yaml.Node]goja.Value{},
		cache:           map[string]*goja.Object{},
		activeCallbacks: map[string]*goja.Object{},
		operationIDs:    map[string]*goja.Object{},
		securitySchemes: map[string]*goja.Object{},
	}
	b.root = b.addDocument(data, uri)
	if b.err != nil {
		return nil, b.err
	}
	model := b.buildModel()
	if b.err != nil {
		return nil, b.err
	}
	freeze, ok := goja.AssertFunction(ctx.runtime.Get("Object").ToObject(ctx.runtime).Get("freeze"))
	if !ok {
		return nil, fmt.Errorf("Object.freeze is unavailable")
	}
	for _, object := range b.objects {
		if _, err := freeze(goja.Undefined(), object); err != nil {
			return nil, fmt.Errorf("freeze OpenAPI model: %w", err)
		}
	}
	return model, nil
}

func (b *apiModelBuilder) fail(format string, args ...any) {
	if b.err == nil {
		b.err = fmt.Errorf(format, args...)
	}
}

func (b *apiModelBuilder) object() *goja.Object {
	o := b.ctx.runtime.NewObject()
	b.objects = append(b.objects, o)
	return o
}

func (b *apiModelBuilder) set(o *goja.Object, name string, value any) {
	if err := o.DefineDataProperty(name, b.ctx.runtime.ToValue(value), goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_TRUE); err != nil {
		b.fail("set %s: %v", name, err)
	}
}

func (b *apiModelBuilder) array(items []*goja.Object) *goja.Object {
	values := make([]any, len(items))
	for i, o := range items {
		values[i] = o
	}
	a := b.ctx.runtime.NewArray(values...)
	b.objects = append(b.objects, a)
	return a
}

func (b *apiModelBuilder) valueArray(items []goja.Value) *goja.Object {
	values := make([]any, len(items))
	for i, v := range items {
		values[i] = v
	}
	a := b.ctx.runtime.NewArray(values...)
	b.objects = append(b.objects, a)
	return a
}

func apiNode(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	return node
}

func apiChild(node *yaml.Node, name string) *yaml.Node {
	node = apiNode(node)
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == name {
			return apiNode(node.Content[i+1])
		}
	}
	return nil
}

func apiString(l apiLocation) string {
	if l.node == nil {
		return ""
	}
	return l.node.Value
}

func apiBool(l apiLocation) bool {
	return l.node != nil && l.node.Tag == "!!bool" && l.node.Value == "true"
}

func apiPointer(parent, key string) string {
	return parent + "/" + strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
}

func (b *apiModelBuilder) child(l apiLocation, name string) apiLocation {
	node := apiChild(l.node, name)
	if node == nil {
		return apiLocation{}
	}
	if loc, ok := b.locations[node]; ok {
		return loc
	}
	l.node = node
	l.pointer = apiPointer(l.pointer, name)
	return l
}

type apiEntry struct {
	name string
	loc  apiLocation
}

func (b *apiModelBuilder) entries(l apiLocation) []apiEntry {
	var result []apiEntry
	if l.node == nil {
		return result
	}
	if l.node.Kind != yaml.MappingNode {
		b.fail("expected object at %s", l.key())
		return result
	}
	for i := 0; i+1 < len(l.node.Content); i += 2 {
		name := l.node.Content[i].Value
		result = append(result, apiEntry{name, b.child(l, name)})
	}
	return result
}

func (b *apiModelBuilder) sequence(l apiLocation) []apiLocation {
	var result []apiLocation
	if !l.exists() {
		return result
	}
	if l.node.Kind != yaml.SequenceNode {
		b.fail("expected array at %s", l.key())
		return result
	}
	for i, node := range l.node.Content {
		loc, ok := b.locations[apiNode(node)]
		if !ok {
			loc = l
			loc.node = apiNode(node)
			loc.pointer = apiPointer(l.pointer, strconv.Itoa(i))
		}
		result = append(result, loc)
	}
	return result
}

func (b *apiModelBuilder) addDocument(data []byte, uri string) apiLocation {
	root, err := parseOpenAPIYAMLDocument(data, uri)
	if err != nil {
		b.fail("%v", err)
		return apiLocation{}
	}
	dialect := ""
	if b.version != "3.0" {
		dialect = "https://spec.openapis.org/oas/" + b.version + "/dialect/base"
	}
	l := apiLocation{node: apiNode(root.Content[0]), documentURI: uri, baseURI: uri, dialect: dialect}
	if b.root.exists() {
		l.dialect = b.root.dialect
	}
	if d := apiChild(l.node, "jsonSchemaDialect"); d != nil {
		l.dialect = d.Value
	}
	if b.version == "3.2" {
		if self := apiChild(l.node, "$self"); self != nil {
			l.baseURI = b.absoluteURI(uri, self.Value)
		}
	}
	b.documents[uri] = l
	b.resources[uri] = l
	// External documents without an OpenAPI header may themselves be schemas.
	b.index(l, apiChild(l.node, "openapi") == nil, "", map[*yaml.Node]bool{})
	return b.locations[l.node]
}

func (b *apiModelBuilder) index(l apiLocation, schema bool, context string, active map[*yaml.Node]bool) {
	if l.node == nil {
		return
	}
	l.node = apiNode(l.node)
	l.schema = schema
	if active[l.node] {
		b.fail("cyclic YAML alias at %s", l.key())
		return
	}
	active[l.node] = true
	defer delete(active, l.node)
	if b.version != "3.0" && schema {
		if id := apiChild(l.node, "$id"); id != nil {
			l.baseURI = b.absoluteURI(l.baseURI, id.Value)
		}
		if d := apiChild(l.node, "$schema"); d != nil {
			l.dialect = d.Value
		}
	}
	b.locations[l.node] = l
	if l.pointer == "" || (schema && apiChild(l.node, "$id") != nil && b.version != "3.0") {
		if old, ok := b.resources[l.baseURI]; ok && old.key() != l.key() {
			b.fail("duplicate schema resource URI %s", l.baseURI)
		}
		b.resources[l.baseURI] = l
	}
	if schema && b.version != "3.0" {
		for _, key := range []string{"$anchor", "$dynamicAnchor"} {
			if a := apiChild(l.node, key); a != nil {
				id := l.baseURI + "#" + a.Value
				if old, ok := b.anchors[id]; ok && old.key() != l.key() {
					b.fail("duplicate schema anchor %s", id)
				}
				b.anchors[id] = l
			}
		}
	}
	if l.node.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i+1 < len(l.node.Content); i += 2 {
			name := l.node.Content[i].Value
			if seen[name] {
				b.fail("duplicate key %q at %s", name, l.key())
			}
			seen[name] = true
			child := l
			child.node = apiNode(l.node.Content[i+1])
			child.pointer = apiPointer(l.pointer, name)
			childSchema := false
			childContext := ""
			if schema {
				switch name {
				case "items", "additionalProperties", "not", "if", "then", "else", "propertyNames", "unevaluatedProperties", "unevaluatedItems", "contains", "contentSchema":
					childSchema = true
				case "properties", "$defs", "patternProperties", "dependentSchemas":
					childContext = "schema-map"
				case "allOf", "oneOf", "anyOf", "prefixItems":
					childContext = "schema-array"
				default:
					childContext = "data"
				}
			} else if context == "schema-map" {
				childSchema = true
			} else if context == "data" || strings.HasPrefix(name, "x-") {
				childContext = "data"
			} else {
				switch name {
				case "schema", "itemSchema":
					childSchema = true
				case "schemas":
					childContext = "schema-map"
				case "example", "examples", "default", "enum", "const", "value", "dataValue":
					// Example payloads contain arbitrary data and do not introduce
					// schema resources, even when their keys resemble schema keywords.
					childContext = "data"
				}
			}
			if !schema && context != "schema-map" && (apiChild(l.node, "operationRef") != nil || (apiChild(l.node, "operationId") != nil && apiChild(l.node, "responses") == nil)) && (name == "parameters" || name == "requestBody") {
				childContext = "data"
				childSchema = false
			}
			b.index(child, childSchema, childContext, active)
		}
	} else if l.node.Kind == yaml.SequenceNode {
		for i, node := range l.node.Content {
			child := l
			child.node = apiNode(node)
			child.pointer = apiPointer(l.pointer, strconv.Itoa(i))
			b.index(child, context == "schema-array", context, active)
		}
	}
}

func (b *apiModelBuilder) absoluteURI(base, ref string) string {
	u, err := url.Parse(base)
	if err != nil {
		b.fail("invalid base URI %q: %v", base, err)
		return ""
	}
	r, err := url.Parse(ref)
	if err != nil {
		b.fail("invalid reference URI %q: %v", ref, err)
		return ""
	}
	return u.ResolveReference(r).String()
}

func (b *apiModelBuilder) resolve(l apiLocation) apiLocation {
	ref := b.child(l, "$ref")
	if !ref.exists() {
		return l
	}
	if ref.node.Tag != "!!str" {
		b.fail("$ref must be a string at %s", ref.key())
		return apiLocation{}
	}
	absolute := b.absoluteURI(l.baseURI, apiString(ref))
	u, err := url.Parse(absolute)
	if err != nil {
		b.fail("invalid reference %q", absolute)
		return apiLocation{}
	}
	fragment := u.Fragment
	u.Fragment = ""
	resourceURI := u.String()
	resource, ok := b.resources[resourceURI]
	if !ok {
		if u.Scheme != "file" || (u.Host != "" && u.Host != "localhost") || u.RawQuery != "" {
			b.fail("cannot load reference %q at %s: only local file references and indexed schema resources are supported", apiString(ref), l.key())
			return apiLocation{}
		}
		data, err := os.ReadFile(filepath.FromSlash(u.Path))
		if err != nil {
			b.fail("read referenced document %s: %v", resourceURI, err)
			return apiLocation{}
		}
		b.addDocument(data, resourceURI)
		resource, ok = b.resources[resourceURI]
		if !ok {
			b.fail("missing reference resource %s", resourceURI)
			return apiLocation{}
		}
	}
	if fragment == "" {
		return resource
	}
	if !strings.HasPrefix(fragment, "/") {
		if target, ok := b.anchors[resourceURI+"#"+fragment]; ok {
			return target
		}
		b.fail("unknown schema anchor %s#%s", resourceURI, fragment)
		return apiLocation{}
	}
	node, err := resolveJSONPointer(resource.node, fragment)
	if err != nil {
		b.fail("resolve reference %q at %s: %v", apiString(ref), l.key(), err)
		return apiLocation{}
	}
	if target, ok := b.locations[apiNode(node)]; ok {
		return target
	}
	b.fail("missing indexed reference target %s", absolute)
	return apiLocation{}
}

func (b *apiModelBuilder) raw(l apiLocation) goja.Value {
	if l.node == nil {
		return goja.Undefined()
	}
	node := apiNode(l.node)
	if value, ok := b.rawValues[node]; ok {
		return value
	}
	if node.Kind == yaml.MappingNode {
		o := b.object()
		b.rawValues[node] = o
		for _, e := range b.entries(l) {
			b.set(o, e.name, b.raw(e.loc))
		}
		return o
	}
	if node.Kind == yaml.SequenceNode {
		a := b.ctx.runtime.NewArray()
		b.objects = append(b.objects, a)
		b.rawValues[node] = a
		for i, item := range b.sequence(l) {
			b.set(a, strconv.Itoa(i), b.raw(item))
		}
		return a
	}
	var data any
	if err := node.Decode(&data); err != nil {
		b.fail("read value at %s: %v", l.key(), err)
		return goja.Undefined()
	}
	// YAML scalars can decode to Go values such as time.Time. Convert through
	// JSON so raw declarations never expose mutable Go wrappers or non-JSON data.
	encoded, err := json.Marshal(data)
	if err != nil {
		b.fail("read JSON value at %s: %v", l.key(), err)
		return goja.Undefined()
	}
	var jsonData any
	if err := json.Unmarshal(encoded, &jsonData); err != nil {
		b.fail("read JSON value at %s: %v", l.key(), err)
		return goja.Undefined()
	}
	value := b.ctx.runtime.ToValue(jsonData)
	b.rawValues[node] = value
	return value
}

func (b *apiModelBuilder) source(l apiLocation) *goja.Object {
	o := b.object()
	kind := "memory"
	if strings.HasPrefix(l.documentURI, "file:") {
		kind = "file"
	}
	b.set(o, "kind", kind)
	b.set(o, "documentUri", l.documentURI)
	b.set(o, "pointer", l.pointer)
	if l.node != nil {
		b.set(o, "line", l.node.Line)
		b.set(o, "column", l.node.Column)
	}
	return o
}

func (b *apiModelBuilder) reference(l, target apiLocation) *goja.Object {
	o := b.object()
	b.set(o, "uri", apiString(b.child(l, "$ref")))
	b.set(o, "resolvedUri", b.absoluteURI(l.baseURI, apiString(b.child(l, "$ref"))))
	b.set(o, "source", b.source(b.child(l, "$ref")))
	b.set(o, "targetSource", b.source(target))
	return o
}

func (b *apiModelBuilder) element(role string, l apiLocation) *goja.Object {
	key := "element:" + role + "@" + l.key()
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	o := b.object()
	b.cache[key] = o
	b.set(o, "id", role+"@"+l.key())
	extensions := b.object()
	if l.exists() {
		b.set(o, "source", b.source(l))
		b.set(o, "raw", b.raw(l))
		for _, e := range b.entriesIfObject(l) {
			if strings.HasPrefix(e.name, "x-") {
				b.set(extensions, e.name, b.raw(e.loc))
			}
		}
	}
	b.set(o, "extensions", extensions)
	return o
}

func (b *apiModelBuilder) entriesIfObject(l apiLocation) []apiEntry {
	if l.node == nil || l.node.Kind != yaml.MappingNode {
		return nil
	}
	return b.entries(l)
}

type apiView struct {
	origin apiLocation
	fields map[string]apiLocation
	order  []string
}

func (v apiView) get(name string) apiLocation { return v.fields[name] }

func (b *apiModelBuilder) view(l apiLocation, pathItem bool, active map[string]bool) apiView {
	v := apiView{origin: l, fields: map[string]apiLocation{}}
	if !l.exists() {
		return v
	}
	if l.node.Kind != yaml.MappingNode {
		b.fail("expected object at %s", l.key())
		return v
	}
	if active[l.key()] {
		b.fail("circular non-schema reference at %s", l.key())
		return v
	}
	if b.child(l, "$ref").exists() {
		active[l.key()] = true
		target := b.resolve(l)
		v = b.view(target, pathItem, active)
		delete(active, l.key())
		v.origin = l
	}
	hasRef := b.child(l, "$ref").exists()
	for _, e := range b.entries(l) {
		if e.name == "$ref" {
			continue
		}
		if hasRef && !pathItem && (b.version == "3.0" || (e.name != "summary" && e.name != "description")) {
			continue
		}
		if pathItem && hasRef {
			if _, exists := v.fields[e.name]; exists {
				b.fail("ambiguous Path Item $ref sibling %q at %s", e.name, l.key())
				continue
			}
		}
		if _, ok := v.fields[e.name]; !ok {
			v.order = append(v.order, e.name)
		}
		v.fields[e.name] = e.loc
	}
	return v
}

func (b *apiModelBuilder) viewElement(role string, v apiView) *goja.Object {
	o := b.element(role, v.origin)
	if b.child(v.origin, "$ref").exists() {
		b.set(o, "reference", b.reference(v.origin, b.resolve(v.origin)))
	}
	return o
}

func (b *apiModelBuilder) copy(o *goja.Object, v apiView, names ...string) {
	for _, name := range names {
		if l := v.get(name); l.exists() {
			b.set(o, name, b.raw(l))
		}
	}
}

func (b *apiModelBuilder) defaultBool(o *goja.Object, v apiView, name string) {
	b.set(o, name, apiBool(v.get(name)))
}

func apiObjectString(o *goja.Object, name string) string {
	v := o.Get(name)
	if v == nil || goja.IsUndefined(v) {
		return ""
	}
	return v.String()
}

func apiStandardMethod(method string, version string) (string, bool) {
	lower := strings.ToLower(method)
	switch lower {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return lower, true
	case "query":
		return lower, version == "3.2"
	}
	return method, false
}

func (b *apiModelBuilder) lookup(o *goja.Object, name string, items []*goja.Object, key string, fold bool) {
	index := map[string]*goja.Object{}
	for _, item := range items {
		if item.Get(key) == nil || goja.IsUndefined(item.Get(key)) {
			continue
		}
		s := apiObjectString(item, key)
		if fold {
			s = strings.ToLower(s)
		}
		index[s] = item
	}
	b.set(o, name, func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		if fold {
			s = strings.ToLower(s)
		}
		if v := index[s]; v != nil {
			return v
		}
		return goja.Undefined()
	})
}

func (b *apiModelBuilder) methodLookup(o *goja.Object, items []*goja.Object) {
	index := map[string]*goja.Object{}
	for _, item := range items {
		index[apiObjectString(item, "method")] = item
	}
	b.set(o, "getOperation", func(call goja.FunctionCall) goja.Value {
		method := call.Argument(0).String()
		if normalized, ok := apiStandardMethod(method, b.version); ok {
			method = normalized
		}
		if v := index[method]; v != nil {
			return v
		}
		return goja.Undefined()
	})
}

func (b *apiModelBuilder) operationLookups(o *goja.Object, items []*goja.Object) {
	index := map[string]map[string]*goja.Object{}
	for _, item := range items {
		path := apiObjectString(item, "path")
		if index[path] == nil {
			index[path] = map[string]*goja.Object{}
		}
		index[path][apiObjectString(item, "method")] = item
	}
	b.set(o, "getOperation", func(call goja.FunctionCall) goja.Value {
		method := call.Argument(0).String()
		if normalized, ok := apiStandardMethod(method, b.version); ok {
			method = normalized
		}
		if v := index[call.Argument(1).String()][method]; v != nil {
			return v
		}
		return goja.Undefined()
	})
	b.lookup(o, "getOperationById", items, "operationId", false)
}

// Parsing a structural copy must not recurse forever on a cyclic YAML alias.
func apiModelValidationBytes(data []byte, uri string) ([]byte, error) {
	document, err := parseOpenAPIYAMLDocument(data, uri)
	if err != nil {
		return nil, err
	}
	var strip func(*yaml.Node, map[*yaml.Node]bool) (*yaml.Node, error)
	strip = func(node *yaml.Node, active map[*yaml.Node]bool) (*yaml.Node, error) {
		node = apiNode(node)
		if node == nil {
			return nil, fmt.Errorf("empty YAML node")
		}
		if active[node] {
			return nil, fmt.Errorf("cyclic YAML alias at line %d", node.Line)
		}
		active[node] = true
		defer delete(active, node)
		clone := cloneYAMLNode(node)
		if node.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(node.Content); i += 2 {
				if node.Content[i].Value == "$ref" {
					continue
				}
				child, err := strip(node.Content[i+1], active)
				if err != nil {
					return nil, err
				}
				clone.Content = append(clone.Content, cloneYAMLTree(node.Content[i]), child)
			}
		} else {
			for _, node := range node.Content {
				child, err := strip(node, active)
				if err != nil {
					return nil, err
				}
				clone.Content = append(clone.Content, child)
			}
		}
		return clone, nil
	}
	node, err := strip(document.Content[0], map[*yaml.Node]bool{})
	if err != nil {
		return nil, err
	}
	return marshalOpenAPINodeJSON(node)
}
