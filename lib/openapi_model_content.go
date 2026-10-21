package ccode

import (
	"strings"

	"github.com/dop251/goja"
)

func (b *apiModelBuilder) parameter(l apiLocation, name string, header bool) *goja.Object {
	role := "parameter"
	if header {
		role = "header:" + name
	}
	key := role + ":" + l.key()
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement(role, v)
	b.cache[key] = o
	if header {
		b.set(o, "name", name)
	} else {
		b.copy(o, v, "name")
		location := apiString(v.get("in"))
		b.set(o, "location", location)
		switch location {
		case "path", "query", "header", "cookie":
		case "querystring":
			if b.version != "3.2" {
				b.fail("querystring parameters require OpenAPI 3.2 at %s", l.key())
			}
		default:
			b.fail("unknown parameter location %q at %s", location, l.key())
		}
		if location == "path" && !apiBool(v.get("required")) {
			b.fail("path parameter %q must be required at %s", apiString(v.get("name")), l.key())
		}
		b.copy(o, v, "allowReserved", "allowEmptyValue")
	}
	b.copy(o, v, "description", "style", "explode", "example")
	b.defaultBool(o, v, "required")
	b.defaultBool(o, v, "deprecated")
	if schema := v.get("schema"); schema.exists() {
		b.set(o, "schema", b.schema(schema))
	}
	content := b.content(v.get("content"))
	b.set(o, "content", b.array(content))
	b.lookup(o, "getMediaType", content, "name", false)
	b.set(o, "examples", b.array(b.examples(v.get("examples"))))
	return o
}

func (b *apiModelBuilder) requestBody(l apiLocation) *goja.Object {
	key := "requestBody:" + l.key()
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement("requestBody", v)
	b.cache[key] = o
	b.copy(o, v, "description")
	b.defaultBool(o, v, "required")
	content := b.content(v.get("content"))
	b.set(o, "content", b.array(content))
	b.lookup(o, "getMediaType", content, "name", false)
	return o
}

func (b *apiModelBuilder) response(l apiLocation, status string) *goja.Object {
	key := "response:" + status + ":" + l.key()
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement("response:"+status, v)
	b.cache[key] = o
	b.set(o, "status", status)
	b.copy(o, v, "description")
	if b.version == "3.2" {
		b.copy(o, v, "summary")
	}
	headers := b.headers(v.get("headers"))
	b.set(o, "headers", b.array(headers))
	b.lookup(o, "getHeader", headers, "name", true)
	content := b.content(v.get("content"))
	b.set(o, "content", b.array(content))
	b.lookup(o, "getMediaType", content, "name", false)
	var links []*goja.Object
	for _, entry := range b.entries(v.get("links")) {
		links = append(links, b.link(entry.loc, entry.name))
	}
	b.set(o, "links", b.array(links))
	b.lookup(o, "getLink", links, "name", false)
	return o
}

func (b *apiModelBuilder) headers(l apiLocation) []*goja.Object {
	var headers []*goja.Object
	seen := map[string]bool{}
	for _, entry := range b.entries(l) {
		key := strings.ToLower(entry.name)
		if seen[key] {
			b.fail("duplicate header %q at %s", entry.name, l.key())
		}
		seen[key] = true
		headers = append(headers, b.parameter(entry.loc, entry.name, true))
	}
	return headers
}

func (b *apiModelBuilder) content(l apiLocation) []*goja.Object {
	var content []*goja.Object
	for _, entry := range b.entries(l) {
		content = append(content, b.mediaType(entry.loc, entry.name))
	}
	return content
}

func (b *apiModelBuilder) mediaType(l apiLocation, name string) *goja.Object {
	key := "mediaType:" + name + ":" + l.key()
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement("mediaType:"+name, v)
	b.cache[key] = o
	b.set(o, "name", name)
	b.copy(o, v, "example")
	if schema := v.get("schema"); schema.exists() {
		b.set(o, "schema", b.schema(schema))
	}
	if b.version == "3.2" {
		if schema := v.get("itemSchema"); schema.exists() {
			b.set(o, "itemSchema", b.schema(schema))
		}
	}
	examples := b.examples(v.get("examples"))
	b.set(o, "examples", b.array(examples))
	b.lookup(o, "getExample", examples, "name", false)
	encodings := b.encodings(v.get("encoding"))
	b.set(o, "encoding", b.array(encodings))
	b.encodingLookup(o, encodings)
	if b.version == "3.2" {
		b.itemEncodings(o, v)
	}
	return o
}

func (b *apiModelBuilder) examples(l apiLocation) []*goja.Object {
	var examples []*goja.Object
	for _, entry := range b.entries(l) {
		v := b.view(entry.loc, false, map[string]bool{})
		o := b.viewElement("example:"+entry.name, v)
		b.set(o, "name", entry.name)
		b.copy(o, v, "summary", "description", "value", "externalValue")
		if b.version == "3.2" {
			b.copy(o, v, "dataValue", "serializedValue")
		}
		examples = append(examples, o)
	}
	return examples
}

func (b *apiModelBuilder) encodings(l apiLocation) []*goja.Object {
	var result []*goja.Object
	for _, entry := range b.entries(l) {
		o := b.object()
		b.set(o, "propertyName", entry.name)
		b.set(o, "encoding", b.encoding(entry.loc))
		result = append(result, o)
	}
	return result
}

func (b *apiModelBuilder) encodingLookup(o *goja.Object, encodings []*goja.Object) {
	index := map[string]goja.Value{}
	for _, encoding := range encodings {
		index[apiObjectString(encoding, "propertyName")] = encoding.Get("encoding")
	}
	b.set(o, "getEncoding", func(call goja.FunctionCall) goja.Value {
		if v := index[call.Argument(0).String()]; v != nil {
			return v
		}
		return goja.Undefined()
	})
}

func (b *apiModelBuilder) itemEncodings(o *goja.Object, v apiView) {
	var prefix []*goja.Object
	for _, item := range b.sequence(v.get("prefixEncoding")) {
		prefix = append(prefix, b.encoding(item))
	}
	b.set(o, "prefixEncoding", b.array(prefix))
	if item := v.get("itemEncoding"); item.exists() {
		b.set(o, "itemEncoding", b.encoding(item))
	}
}

func (b *apiModelBuilder) encoding(l apiLocation) *goja.Object {
	key := "encoding:" + l.key()
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement("encoding", v)
	b.cache[key] = o
	b.copy(o, v, "contentType", "style", "explode", "allowReserved")
	b.set(o, "headers", b.array(b.headers(v.get("headers"))))
	if b.version == "3.2" {
		b.set(o, "encoding", b.array(b.encodings(v.get("encoding"))))
		b.itemEncodings(o, v)
	}
	return o
}

func (b *apiModelBuilder) link(l apiLocation, name string) *goja.Object {
	v := b.view(l, false, map[string]bool{})
	o := b.viewElement("link:"+name, v)
	b.set(o, "name", name)
	b.copy(o, v, "operationId", "operationRef", "description", "requestBody")
	if params := v.get("parameters"); params.exists() {
		b.set(o, "parameters", b.raw(params))
	} else {
		b.set(o, "parameters", b.object())
	}
	if server := v.get("server"); server.exists() {
		b.set(o, "server", b.server(server, "operation", false, "link:"+name))
	}
	return o
}
