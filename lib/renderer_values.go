package ccode

import (
	"fmt"
	"strconv"

	"github.com/dop251/goja"
	"github.com/nikolalohinski/gonja/v2/exec"
)

// templateValueBridge belongs to one rendering operation. It eagerly converts
// data to Gonja's existing ordered dictionaries and lists, without copying a
// shared object twice. Callable results use the same cache as the initial data.
// Like the owning Goja runtime, a bridge is used synchronously on one goroutine.
type templateValueBridge struct {
	runtime      *goja.Runtime
	objects      map[*goja.Object]any
	functions    map[templateFunctionKey]func(*exec.VarArgs) (*exec.Value, error)
	dictionaries map[**exec.Pair]*goja.Object
	arrays       map[templateArrayKey]*goja.Object
}

type templateFunctionKey struct{ function, receiver *goja.Object }

// Include length so a slice starting at index zero is not mistaken for its
// original, longer JavaScript array when passed back to a method.
type templateArrayKey struct {
	first  *any
	length int
}

func gojaValueToTemplateData(runtime *goja.Runtime, value goja.Value) (any, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runner runtime is not initialized")
	}
	bridge := &templateValueBridge{
		runtime:      runtime,
		objects:      map[*goja.Object]any{},
		functions:    map[templateFunctionKey]func(*exec.VarArgs) (*exec.Value, error){},
		dictionaries: map[**exec.Pair]*goja.Object{},
		arrays:       map[templateArrayKey]*goja.Object{},
	}
	return bridge.convert(value, nil)
}

// convert catches JavaScript exceptions from getters as well as conversions,
// so they reach the caller as rendering errors rather than Go panics.
func (b *templateValueBridge) convert(value goja.Value, receiver *goja.Object) (result any, err error) {
	if exception := b.runtime.Try(func() { result, err = b.convertValue(value, receiver) }); exception != nil {
		return nil, exception
	}
	return result, err
}

func (b *templateValueBridge) convertValue(value goja.Value, receiver *goja.Object) (any, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, nil
	}
	object, ok := value.(*goja.Object)
	if !ok {
		return value.Export(), nil
	}
	if function, ok := goja.AssertFunction(object); ok {
		return b.callable(object, receiver, function), nil
	}
	if cached, ok := b.objects[object]; ok {
		return cached, nil
	}
	switch object.ClassName() {
	case "Array":
		length := int(object.Get("length").ToInteger())
		// A spare backing slot gives distinct empty arrays an identity too.
		items := make([]any, length, max(1, length))
		b.objects[object] = items
		b.arrays[templateArrayKey{&items[:cap(items)][0], len(items)}] = object
		for i := range items {
			item, err := b.convertValue(object.Get(strconv.Itoa(i)), object)
			if err != nil {
				return nil, fmt.Errorf("convert array index %d: %w", i, err)
			}
			items[i] = item
		}
		return items, nil
	case "Object":
		keys := object.Keys()
		// Cache a complete-length slice before descending. Copies of exec.Dict
		// share that slice, so back-references see the populated dictionary later.
		dict := exec.Dict{Pairs: make([]*exec.Pair, len(keys), max(1, len(keys)))}
		b.objects[object] = dict
		b.dictionaries[&dict.Pairs[:cap(dict.Pairs)][0]] = object
		for i, key := range keys {
			item, err := b.convertValue(object.Get(key), object)
			if err != nil {
				return nil, fmt.Errorf("convert object key %q: %w", key, err)
			}
			dict.Pairs[i] = &exec.Pair{Key: exec.AsValue(key), Value: exec.AsValue(item)}
		}
		return dict, nil
	default:
		return object.Export(), nil
	}
}

func (b *templateValueBridge) callable(object, receiver *goja.Object, function goja.Callable) func(*exec.VarArgs) (*exec.Value, error) {
	key := templateFunctionKey{object, receiver}
	if cached := b.functions[key]; cached != nil {
		return cached
	}
	wrapper := func(arguments *exec.VarArgs) (*exec.Value, error) {
		if len(arguments.KwArgs) != 0 {
			return nil, fmt.Errorf("template calls to JavaScript functions accept positional arguments only; pass options as a dictionary")
		}
		input := templateCallArguments{bridge: b, dictionaries: map[**exec.Pair]*goja.Object{}, arrays: map[templateArrayKey]*goja.Object{}, lists: map[**exec.Value]*goja.Object{}}
		args := make([]goja.Value, len(arguments.Args))
		for i, arg := range arguments.Args {
			converted, err := input.convert(arg)
			if err != nil {
				return nil, fmt.Errorf("convert function argument %d: %w", i+1, err)
			}
			args[i] = converted
		}
		var this goja.Value = goja.Undefined()
		if receiver != nil {
			this = receiver
		}
		result, err := function(this, args...)
		if err != nil {
			return nil, fmt.Errorf("call template function: %w", err)
		}
		converted, err := b.convert(result, nil)
		if err != nil {
			return nil, fmt.Errorf("convert function result: %w", err)
		}
		return exec.AsValue(converted), nil
	}
	b.functions[key] = wrapper
	return wrapper
}

// Return existing JavaScript objects when a template passes model objects back
// to a method. Only newly constructed template dictionaries/lists are copied.
// Separate per-call caches also preserve aliases/cycles in those arguments.
type templateCallArguments struct {
	bridge       *templateValueBridge
	dictionaries map[**exec.Pair]*goja.Object
	arrays       map[templateArrayKey]*goja.Object
	lists        map[**exec.Value]*goja.Object
}

func (a *templateCallArguments) convert(value *exec.Value) (goja.Value, error) {
	if value == nil || value.IsNil() {
		return goja.Null(), nil
	}
	switch data := value.Interface().(type) {
	case exec.Dict:
		return a.dictionary(data)
	case *exec.Dict:
		return a.dictionary(*data)
	case []any:
		var key templateArrayKey
		if cap(data) > 0 {
			key = templateArrayKey{&data[:cap(data)][0], len(data)}
			if original := a.bridge.arrays[key]; original != nil {
				return original, nil
			}
			if cached := a.arrays[key]; cached != nil {
				return cached, nil
			}
		}
		array := a.bridge.runtime.NewArray()
		if key.first != nil {
			a.arrays[key] = array
		}
		for i, item := range data {
			if err := a.setArrayItem(array, i, exec.AsValue(item)); err != nil {
				return nil, err
			}
		}
		return array, nil
	case exec.ValuesList:
		var key **exec.Value
		if cap(data) > 0 {
			key = &data[:cap(data)][0]
			if cached := a.lists[key]; cached != nil {
				return cached, nil
			}
		}
		array := a.bridge.runtime.NewArray()
		if key != nil {
			a.lists[key] = array
		}
		for i, item := range data {
			if err := a.setArrayItem(array, i, item); err != nil {
				return nil, err
			}
		}
		return array, nil
	default:
		// Gonja literals above require explicit conversion. Its primitive values
		// and native Go filter results can use Goja's standard Go-value conversion.
		return a.bridge.runtime.ToValue(data), nil
	}
}

func (a *templateCallArguments) dictionary(dict exec.Dict) (goja.Value, error) {
	var key **exec.Pair
	if cap(dict.Pairs) > 0 {
		key = &dict.Pairs[:cap(dict.Pairs)][0]
		if original := a.bridge.dictionaries[key]; original != nil {
			return original, nil
		}
		if cached := a.dictionaries[key]; cached != nil {
			return cached, nil
		}
	}
	object := a.bridge.runtime.NewObject()
	if key != nil {
		a.dictionaries[key] = object
	}
	for _, pair := range dict.Pairs {
		if !pair.Key.IsString() {
			return nil, fmt.Errorf("JavaScript options dictionaries require string keys")
		}
		value, err := a.convert(pair.Value)
		if err != nil {
			return nil, err
		}
		// Define an own property; keys such as __proto__ are data, not setters.
		if err := object.DefineDataProperty(pair.Key.String(), value, goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_TRUE); err != nil {
			return nil, err
		}
	}
	return object, nil
}

func (a *templateCallArguments) setArrayItem(array *goja.Object, index int, item *exec.Value) error {
	converted, err := a.convert(item)
	if err != nil {
		return err
	}
	return array.Set(strconv.Itoa(index), converted)
}
