package ccode

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

type templateFilterSpec struct {
	name   string
	filter exec.FilterFunction
}

var templateFilterSpecs = []templateFilterSpec{
	{name: "camelCase", filter: initialismTemplateFilter("camelCase", stringToCamelCase)},
	{name: "pascalCase", filter: initialismTemplateFilter("pascalCase", stringToPascalCase)},
	{name: "snakeCase", filter: stringTemplateFilter("snakeCase", stringToSnakeCase)},
	{name: "kebabCase", filter: stringTemplateFilter("kebabCase", stringToKebabCase)},
	{name: "constantCase", filter: stringTemplateFilter("constantCase", stringToConstantCase)},
	{name: "dotCase", filter: stringTemplateFilter("dotCase", stringToDotCase)},
	{name: "pathCase", filter: stringTemplateFilter("pathCase", stringToPathCase)},
	{name: "titleCase", filter: initialismTemplateFilter("titleCase", stringToTitleCase)},
	{name: "sentenceCase", filter: initialismTemplateFilter("sentenceCase", stringToSentenceCase)},
	{name: "upperFirst", filter: stringTemplateFilter("upperFirst", stringToUpperFirst)},
	{name: "lowerFirst", filter: stringTemplateFilter("lowerFirst", stringToLowerFirst)},
	{name: "normalizeSpace", filter: stringTemplateFilter("normalizeSpace", stringToNormalizeSpace)},
	{name: "goExported", filter: initialismTemplateFilter("goExported", stringToGoExported)},
	{name: "goUnexported", filter: initialismTemplateFilter("goUnexported", stringToGoUnexported)},
	{name: "goPackage", filter: stringTemplateFilter("goPackage", stringToGoPackage)},
}

var (
	templateEnvironmentOnce sync.Once
	templateEnvironment     *exec.Environment
	templateEnvironmentErr  error
)

func cohesiveTemplateEnvironment() (*exec.Environment, error) {
	// Construction and registration finish before the environment is published.
	// Rendering only reads the environment and its filter set after this point.
	templateEnvironmentOnce.Do(func() {
		templateEnvironment, templateEnvironmentErr = newCohesiveTemplateEnvironment()
	})

	return templateEnvironment, templateEnvironmentErr
}

func newCohesiveTemplateEnvironment() (*exec.Environment, error) {
	defaultEnvironment := gonja.DefaultEnvironment
	// FilterSet.Update copies the registered functions into the new set. A shallow
	// Environment copy would retain Gonja's shared, mutable default FilterSet.
	filters := exec.NewFilterSet(map[string]exec.FilterFunction{}).Update(defaultEnvironment.Filters)

	if err := registerTemplateFilters(filters); err != nil {
		return nil, err
	}

	return &exec.Environment{
		Filters:           filters,
		ControlStructures: defaultEnvironment.ControlStructures,
		Tests:             defaultEnvironment.Tests,
		Context:           defaultEnvironment.Context,
		Methods:           defaultEnvironment.Methods,
	}, nil
}

func registerTemplateFilters(filters *exec.FilterSet) error {
	for _, spec := range templateFilterSpecs {
		if err := filters.Register(spec.name, spec.filter); err != nil {
			return fmt.Errorf("register Cohesive Code template filter %q: %w", spec.name, err)
		}
	}

	return nil
}

func stringTemplateFilter(name string, transform func(string) string) exec.FilterFunction {
	return func(_ *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
		if err := validateTemplateFilterArguments(params); err != nil {
			return templateFilterError(name, err)
		}

		input, err := templateFilterString(in)
		if err != nil {
			return templateFilterError(name, err)
		}

		return exec.AsValue(transform(input))
	}
}

func initialismTemplateFilter(name string, transform func(string, []string) string) exec.FilterFunction {
	return func(_ *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
		initialisms, err := templateFilterInitialisms(params)
		if err != nil {
			return templateFilterError(name, err)
		}

		input, err := templateFilterString(in)
		if err != nil {
			return templateFilterError(name, err)
		}

		return exec.AsValue(transform(input, initialisms))
	}
}

func validateTemplateFilterArguments(params *exec.VarArgs, allowedKeywords ...string) error {
	if params == nil {
		return nil
	}
	if len(params.Args) != 0 {
		return fmt.Errorf("positional arguments are not supported")
	}

	allowed := make(map[string]struct{}, len(allowedKeywords))
	for _, keyword := range allowedKeywords {
		allowed[keyword] = struct{}{}
	}

	unexpected := make([]string, 0)
	for keyword := range params.KwArgs {
		if _, ok := allowed[keyword]; !ok {
			unexpected = append(unexpected, keyword)
		}
	}
	if len(unexpected) == 0 {
		return nil
	}

	sort.Strings(unexpected)
	if len(unexpected) == 1 {
		return fmt.Errorf("unexpected argument %q", unexpected[0])
	}

	quoted := make([]string, 0, len(unexpected))
	for _, keyword := range unexpected {
		quoted = append(quoted, fmt.Sprintf("%q", keyword))
	}
	return fmt.Errorf("unexpected arguments %s", strings.Join(quoted, ", "))
}

func templateFilterInitialisms(params *exec.VarArgs) ([]string, error) {
	if err := validateTemplateFilterArguments(params, "initialisms"); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, nil
	}

	value, supplied := params.KwArgs["initialisms"]
	if !supplied {
		return nil, nil
	}
	if value == nil || !value.IsList() {
		return nil, fmt.Errorf("%q must be a list of strings", "initialisms")
	}

	initialisms := make([]string, 0, value.Len())
	var validationErr error
	value.Iterate(func(index, _ int, item, _ *exec.Value) bool {
		if item == nil || !item.IsString() {
			validationErr = fmt.Errorf("initialisms[%d] must be a string", index)
			return false
		}
		if isStringBlank(item.String()) {
			validationErr = fmt.Errorf("initialisms[%d] must not be blank", index)
			return false
		}

		initialisms = append(initialisms, item.String())
		return true
	}, func() {})
	if validationErr != nil {
		return nil, validationErr
	}

	return initialisms, nil
}

func templateFilterString(value *exec.Value) (string, error) {
	if value == nil || !value.IsString() {
		return "", fmt.Errorf("input must be a string")
	}

	return value.String(), nil
}

func templateFilterError(name string, err error) *exec.Value {
	return exec.AsValue(exec.ErrInvalidCall(fmt.Errorf("filter %q: %w", name, err)))
}
