package ccode

import (
	"github.com/dop251/goja"
	"gopkg.in/yaml.v3"
)

func (b *apiModelBuilder) namedSchemas(l apiLocation) []*goja.Object {
	var result []*goja.Object
	for _, entry := range b.entries(l) {
		o := b.element("namedSchema:"+entry.name, entry.loc)
		b.set(o, "name", entry.name)
		b.set(o, "schema", b.schema(entry.loc))
		if b.child(entry.loc, "$ref").exists() {
			b.set(o, "reference", b.reference(entry.loc, b.resolve(entry.loc)))
		}
		result = append(result, o)
	}
	return result
}

func (b *apiModelBuilder) schema(l apiLocation) *goja.Object { return b.schemaValue(l, false) }

func (b *apiModelBuilder) schemaValue(l apiLocation, allowBoolean bool) *goja.Object {
	if !l.exists() {
		b.fail("missing schema declaration")
		return b.object()
	}
	// A fragment in a standalone external document may be a schema even if
	// its containing object is merely a map of definitions.
	if !l.schema {
		b.index(l, true, "", map[*yaml.Node]bool{})
		l = b.locations[l.node]
	}
	if l.node.Tag == "!!bool" && b.version == "3.0" && !allowBoolean {
		b.fail("boolean schemas require OpenAPI 3.1 or later at %s", l.key())
	}
	key := "schema:" + l.key()
	if cached := b.cache[key]; cached != nil {
		return cached
	}
	o := b.element("schema", l)
	b.cache[key] = o
	if l.node.Tag == "!!bool" {
		b.set(o, "kind", "boolean")
		b.set(o, "value", apiBool(l))
		return o
	}
	if l.node.Kind != yaml.MappingNode {
		b.fail("schema must be an object or boolean at %s", l.key())
		return o
	}
	b.set(o, "kind", "schema")
	modern := b.version != "3.0"
	// 3.0 Reference Objects ignore siblings. Modern JSON Schema $ref is an
	// applicator: its target and local sibling constraints are kept separately.
	local := l
	if b.child(l, "$ref").exists() {
		target := b.resolve(l)
		ref := b.reference(l, target)
		b.set(ref, "target", b.schema(target))
		b.set(o, "reference", ref)
		if !modern {
			local = apiLocation{}
		}
	}
	v := apiView{origin: local, fields: map[string]apiLocation{}}
	for _, entry := range b.entriesIfObject(local) {
		v.fields[entry.name] = entry.loc
		v.order = append(v.order, entry.name)
	}
	b.copy(o, v, "title", "description", "format", "enum", "default", "example", "readOnly", "writeOnly", "deprecated", "multipleOf", "minimum", "maximum", "minLength", "maxLength", "pattern", "minItems", "maxItems", "uniqueItems", "minProperties", "maxProperties")
	for _, name := range []string{"exclusiveMinimum", "exclusiveMaximum"} {
		if value := v.get(name); value.exists() {
			if modern && value.node.Tag != "!!int" && value.node.Tag != "!!float" {
				b.fail("%s must be numeric in OpenAPI %s at %s", name, b.version, value.key())
			}
			if !modern && value.node.Tag != "!!bool" {
				b.fail("%s must be boolean in OpenAPI 3.0 at %s", name, value.key())
			}
			b.set(o, name, b.raw(value))
		}
	}
	var types []goja.Value
	if declared := v.get("type"); declared.exists() {
		values := []apiLocation{declared}
		if declared.node.Kind == yaml.SequenceNode {
			values = b.sequence(declared)
			if !modern {
				b.fail("type arrays require OpenAPI 3.1 or later at %s", declared.key())
			}
		}
		for _, value := range values {
			switch apiString(value) {
			case "object", "array", "string", "integer", "number", "boolean":
			case "null":
				if !modern {
					b.fail("null type requires OpenAPI 3.1 or later at %s", value.key())
				}
			default:
				b.fail("unknown schema type %q at %s", apiString(value), value.key())
			}
			types = append(types, b.raw(value))
		}
	}
	b.set(o, "types", b.valueArray(types))
	required := v.get("required")
	if required.exists() {
		b.set(o, "required", b.raw(required))
	} else {
		b.set(o, "required", b.array(nil))
	}
	requiredNames := map[string]bool{}
	for _, name := range b.sequence(required) {
		requiredNames[apiString(name)] = true
	}
	var properties []*goja.Object
	for _, entry := range b.entries(v.get("properties")) {
		property := b.element("property:"+entry.name, entry.loc)
		b.set(property, "name", entry.name)
		b.set(property, "parent", o)
		b.set(property, "schema", b.schema(entry.loc))
		b.set(property, "required", requiredNames[entry.name])
		properties = append(properties, property)
	}
	b.set(o, "properties", b.array(properties))
	b.lookup(o, "getProperty", properties, "name", false)
	for _, keyword := range []string{"items", "not", "additionalProperties"} {
		if value := v.get(keyword); value.exists() {
			b.set(o, keyword, b.schemaValue(value, keyword == "additionalProperties"))
		}
	}
	for _, keyword := range []string{"allOf", "oneOf", "anyOf"} {
		b.set(o, keyword, b.schemaArray(v.get(keyword)))
	}
	if modern {
		b.set(o, "dialect", l.dialect)
		for _, pair := range [][2]string{{"$id", "schemaId"}, {"$anchor", "anchor"}, {"$dynamicAnchor", "dynamicAnchor"}, {"$dynamicRef", "dynamicRef"}} {
			if value := v.get(pair[0]); value.exists() {
				b.set(o, pair[1], b.raw(value))
			}
		}
		b.copy(o, v, "const", "minContains", "maxContains", "contentEncoding", "contentMediaType")
		defs := b.namedSchemas(v.get("$defs"))
		b.set(o, "definitions", b.array(defs))
		b.lookup(o, "getDefinition", defs, "name", false)
		b.set(o, "patternProperties", b.array(b.namedSchemas(v.get("patternProperties"))))
		b.set(o, "dependentSchemas", b.array(b.namedSchemas(v.get("dependentSchemas"))))
		if value := v.get("dependentRequired"); value.exists() {
			b.set(o, "dependentRequired", b.raw(value))
		} else {
			b.set(o, "dependentRequired", b.object())
		}
		b.set(o, "prefixItems", b.schemaArray(v.get("prefixItems")))
		if examples := v.get("examples"); examples.exists() {
			b.set(o, "examples", b.raw(examples))
		} else {
			b.set(o, "examples", b.array(nil))
		}
		for _, keyword := range []string{"if", "then", "else", "propertyNames", "unevaluatedProperties", "unevaluatedItems", "contains", "contentSchema"} {
			if value := v.get(keyword); value.exists() {
				b.set(o, keyword, b.schema(value))
			}
		}
	} else {
		b.copy(o, v, "nullable")
	}
	if value := v.get("discriminator"); value.exists() {
		disc := b.simple(value, "discriminator", "propertyName")
		if mapping := b.child(value, "mapping"); mapping.exists() {
			b.set(disc, "mapping", b.raw(mapping))
		} else {
			b.set(disc, "mapping", b.object())
		}
		if b.version == "3.2" {
			if mapping := b.child(value, "defaultMapping"); mapping.exists() {
				b.set(disc, "defaultMapping", b.raw(mapping))
			}
		}
		b.set(o, "discriminator", disc)
	}
	if value := v.get("xml"); value.exists() {
		names := []string{"name", "namespace", "prefix", "attribute", "wrapped"}
		if b.version == "3.2" {
			names = append(names, "nodeType")
		}
		b.set(o, "xml", b.simple(value, "xml", names...))
	}
	if value := v.get("externalDocs"); value.exists() {
		b.set(o, "externalDocs", b.externalDocs(value))
	}
	return o
}

func (b *apiModelBuilder) schemaArray(l apiLocation) *goja.Object {
	var result []*goja.Object
	for _, entry := range b.sequence(l) {
		result = append(result, b.schema(entry))
	}
	return b.array(result)
}
