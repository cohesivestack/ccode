package ccode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type openAPIReferenceResolver struct {
	documents map[string]*yaml.Node
}

func resolveOpenAPIFileReferences(specBytes []byte, rootPath string) ([]byte, error) {
	rootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("resolve root file path: %w", err)
	}
	rootPath = filepath.Clean(rootPath)

	root, err := parseOpenAPIYAMLDocument(specBytes, rootPath)
	if err != nil {
		return nil, err
	}

	resolver := &openAPIReferenceResolver{
		documents: map[string]*yaml.Node{rootPath: root},
	}
	resolved, err := resolver.resolveNode(root.Content[0], rootPath, make(map[string]bool))
	if err != nil {
		return nil, err
	}

	return marshalOpenAPINodeJSON(resolved)
}

func parseOpenAPIYAMLDocument(documentBytes []byte, filePath string) (*yaml.Node, error) {
	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(documentBytes))
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("parse referenced OpenAPI document %s: %w", filePath, err)
	}
	if len(document.Content) != 1 {
		return nil, fmt.Errorf("parse referenced OpenAPI document %s: document is empty", filePath)
	}
	return &document, nil
}

func (r *openAPIReferenceResolver) resolveNode(node *yaml.Node, sourcePath string, active map[string]bool) (*yaml.Node, error) {
	if node == nil {
		return nil, nil
	}
	if node.Kind == yaml.AliasNode {
		return r.resolveNode(node.Alias, sourcePath, active)
	}

	if node.Kind == yaml.MappingNode {
		refValue, hasRef, err := mappingReference(node)
		if err != nil {
			return nil, fmt.Errorf("in %s: %w", sourcePath, err)
		}
		if hasRef {
			return r.resolveReferenceMapping(node, refValue, sourcePath, active)
		}
	}

	resolved := cloneYAMLNode(node)
	resolved.Content = nil
	for _, child := range node.Content {
		resolvedChild, err := r.resolveNode(child, sourcePath, active)
		if err != nil {
			return nil, err
		}
		resolved.Content = append(resolved.Content, resolvedChild)
	}
	return resolved, nil
}

func mappingReference(node *yaml.Node) (string, bool, error) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value != "$ref" {
			continue
		}
		if node.Content[i+1].Kind != yaml.ScalarNode || node.Content[i+1].Tag != "!!str" {
			return "", false, fmt.Errorf("$ref must be a string")
		}
		return node.Content[i+1].Value, true, nil
	}
	return "", false, nil
}

func (r *openAPIReferenceResolver) resolveReferenceMapping(node *yaml.Node, reference, sourcePath string, active map[string]bool) (*yaml.Node, error) {
	targetPath, pointer, err := resolveOpenAPIReferenceLocation(reference, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("resolve reference %q from %s: %w", reference, sourcePath, err)
	}

	referenceID := targetPath + "#" + pointer
	if active[referenceID] {
		return cloneYAMLTree(node), nil
	}

	targetDocument, err := r.loadDocument(targetPath)
	if err != nil {
		return nil, fmt.Errorf("resolve reference %q from %s: %w", reference, sourcePath, err)
	}
	target, err := resolveJSONPointer(targetDocument.Content[0], pointer)
	if err != nil {
		return nil, fmt.Errorf("resolve reference %q from %s: %w", reference, sourcePath, err)
	}
	if target.Kind == yaml.ScalarNode && target.Tag == "!!bool" {
		return cloneYAMLTree(target), nil
	}
	if target.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("resolve reference %q from %s: incompatible referenced value at #%s: expected an object", reference, sourcePath, pointer)
	}

	active[referenceID] = true
	resolvedTarget, err := r.resolveNode(target, targetPath, active)
	delete(active, referenceID)
	if err != nil {
		return nil, err
	}

	return r.mergeReference(node, resolvedTarget, sourcePath, active)
}

func (r *openAPIReferenceResolver) mergeReference(source, target *yaml.Node, sourcePath string, active map[string]bool) (*yaml.Node, error) {
	result := cloneYAMLNode(source)
	result.Content = nil

	for i := 0; i+1 < len(source.Content); i += 2 {
		if source.Content[i].Value == "$ref" {
			result.Content = append(result.Content, cloneYAMLTree(source.Content[i]), cloneYAMLTree(source.Content[i+1]))
			break
		}
	}

	for i := 0; i+1 < len(target.Content); i += 2 {
		if target.Content[i].Value == "$ref" {
			continue
		}
		result.Content = append(result.Content, cloneYAMLTree(target.Content[i]), cloneYAMLTree(target.Content[i+1]))
	}

	for i := 0; i+1 < len(source.Content); i += 2 {
		key := source.Content[i]
		if key.Value == "$ref" {
			continue
		}
		value, err := r.resolveNode(source.Content[i+1], sourcePath, active)
		if err != nil {
			return nil, err
		}
		if replaceMappingValue(result, key.Value, value) {
			continue
		}
		result.Content = append(result.Content, cloneYAMLTree(key), value)
	}

	return result, nil
}

func replaceMappingValue(mapping *yaml.Node, key string, value *yaml.Node) bool {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return true
		}
	}
	return false
}

func (r *openAPIReferenceResolver) loadDocument(filePath string) (*yaml.Node, error) {
	filePath = filepath.Clean(filePath)
	if document, ok := r.documents[filePath]; ok {
		return document, nil
	}

	documentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read referenced OpenAPI file %s: %w", filePath, err)
	}
	document, err := parseOpenAPIYAMLDocument(documentBytes, filePath)
	if err != nil {
		return nil, err
	}
	r.documents[filePath] = document
	return document, nil
}

func resolveOpenAPIReferenceLocation(reference, sourcePath string) (string, string, error) {
	referenceURL, err := url.Parse(reference)
	if err != nil {
		return "", "", fmt.Errorf("invalid reference URI: %w", err)
	}
	if referenceURL.Scheme != "" || referenceURL.Host != "" {
		return "", "", fmt.Errorf("unsupported non-file reference URI")
	}
	if referenceURL.RawQuery != "" {
		return "", "", fmt.Errorf("file reference must not contain a query")
	}

	referencePath, err := url.PathUnescape(referenceURL.EscapedPath())
	if err != nil {
		return "", "", fmt.Errorf("invalid escaped file path: %w", err)
	}
	targetPath := sourcePath
	if referencePath != "" {
		if filepath.IsAbs(referencePath) {
			targetPath = filepath.Clean(referencePath)
		} else {
			targetPath = filepath.Clean(filepath.Join(filepath.Dir(sourcePath), filepath.FromSlash(referencePath)))
		}
	}

	pointer, err := url.PathUnescape(referenceURL.EscapedFragment())
	if err != nil {
		return "", "", fmt.Errorf("invalid escaped fragment: %w", err)
	}
	if pointer != "" && !strings.HasPrefix(pointer, "/") {
		return "", "", fmt.Errorf("invalid JSON Pointer #%s: pointer must be empty or start with /", pointer)
	}
	return targetPath, pointer, nil
}

func resolveJSONPointer(root *yaml.Node, pointer string) (*yaml.Node, error) {
	current := root
	if pointer == "" {
		return current, nil
	}

	for _, encodedToken := range strings.Split(pointer[1:], "/") {
		token, err := decodeJSONPointerToken(encodedToken)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON Pointer #%s: %w", pointer, err)
		}
		if current.Kind == yaml.AliasNode {
			current = current.Alias
		}
		switch current.Kind {
		case yaml.MappingNode:
			var next *yaml.Node
			for i := 0; i+1 < len(current.Content); i += 2 {
				if current.Content[i].Value == token {
					next = current.Content[i+1]
					break
				}
			}
			if next == nil {
				return nil, fmt.Errorf("JSON Pointer #%s does not exist (missing %q)", pointer, token)
			}
			current = next
		case yaml.SequenceNode:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(current.Content) {
				return nil, fmt.Errorf("JSON Pointer #%s has invalid array index %q", pointer, token)
			}
			current = current.Content[index]
		default:
			return nil, fmt.Errorf("JSON Pointer #%s traverses through a non-container at %q", pointer, token)
		}
	}
	return current, nil
}

func decodeJSONPointerToken(token string) (string, error) {
	var decoded strings.Builder
	for i := 0; i < len(token); i++ {
		if token[i] != '~' {
			decoded.WriteByte(token[i])
			continue
		}
		if i+1 >= len(token) {
			return "", fmt.Errorf("invalid escape in token %q", token)
		}
		i++
		switch token[i] {
		case '0':
			decoded.WriteByte('~')
		case '1':
			decoded.WriteByte('/')
		default:
			return "", fmt.Errorf("invalid escape ~%c in token %q", token[i], token)
		}
	}
	return decoded.String(), nil
}

func cloneYAMLTree(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	clone := cloneYAMLNode(node)
	clone.Content = make([]*yaml.Node, 0, len(node.Content))
	for _, child := range node.Content {
		clone.Content = append(clone.Content, cloneYAMLTree(child))
	}
	if node.Alias != nil {
		clone.Alias = cloneYAMLTree(node.Alias)
	}
	return clone
}

func cloneYAMLNode(node *yaml.Node) *yaml.Node {
	clone := *node
	clone.Content = nil
	clone.Alias = nil
	return &clone
}

func marshalOpenAPINodeJSON(node *yaml.Node) ([]byte, error) {
	var output bytes.Buffer
	if err := writeOpenAPINodeJSON(&output, node); err != nil {
		return nil, fmt.Errorf("render resolved OpenAPI document as JSON: %w", err)
	}
	return output.Bytes(), nil
}

func stripMaterializedOpenAPIReferences(documentBytes []byte, source string) ([]byte, error) {
	document, err := parseOpenAPIYAMLDocument(documentBytes, source)
	if err != nil {
		return nil, err
	}
	stripped := stripOpenAPIReferenceNodes(document.Content[0])
	return marshalOpenAPINodeJSON(stripped)
}

func stripOpenAPIReferenceNodes(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.AliasNode {
		return stripOpenAPIReferenceNodes(node.Alias)
	}
	stripped := cloneYAMLNode(node)
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == "$ref" {
				continue
			}
			stripped.Content = append(stripped.Content, cloneYAMLTree(node.Content[i]), stripOpenAPIReferenceNodes(node.Content[i+1]))
		}
		return stripped
	}
	for _, child := range node.Content {
		stripped.Content = append(stripped.Content, stripOpenAPIReferenceNodes(child))
	}
	return stripped
}

func writeOpenAPINodeJSON(output *bytes.Buffer, node *yaml.Node) error {
	if node.Kind == yaml.AliasNode {
		return writeOpenAPINodeJSON(output, node.Alias)
	}
	switch node.Kind {
	case yaml.MappingNode:
		output.WriteByte('{')
		for i := 0; i+1 < len(node.Content); i += 2 {
			if i > 0 {
				output.WriteByte(',')
			}
			if node.Content[i].Kind != yaml.ScalarNode {
				return fmt.Errorf("object key at line %d must be a scalar", node.Content[i].Line)
			}
			key, err := json.Marshal(node.Content[i].Value)
			if err != nil {
				return err
			}
			output.Write(key)
			output.WriteByte(':')
			if err := writeOpenAPINodeJSON(output, node.Content[i+1]); err != nil {
				return err
			}
		}
		output.WriteByte('}')
	case yaml.SequenceNode:
		output.WriteByte('[')
		for i, child := range node.Content {
			if i > 0 {
				output.WriteByte(',')
			}
			if err := writeOpenAPINodeJSON(output, child); err != nil {
				return err
			}
		}
		output.WriteByte(']')
	case yaml.ScalarNode:
		var value any
		if err := node.Decode(&value); err != nil {
			return err
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return err
		}
		output.Write(encoded)
	default:
		return fmt.Errorf("unsupported YAML node kind %d at line %d", node.Kind, node.Line)
	}
	return nil
}
