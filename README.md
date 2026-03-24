# Cohesive Code

Cohesive Code is an AI‑enabled code generator.

This intent will evolve in different stages:

## Stage 1 - Add a template generator with support from OpenAPI models:
  * A CLI command which is able to run process
  * Process are typescript functions executed from the CLI or from another process
  * Templates which are Jinja template and are called from a Process
  * A process has a typescript API:
    * The object ccode is the package that has the API to communicate with the model and API offer by the Cohesive Code library
    * The type `ccode.OpenAPI` is an openAPI model
    * The function `ccode.GetOpenAPI(filePath: string): ccode.OpenAPI` parse a yaml or json file to an OpenAPI model
    * The function `ccode.TemplateToString(templatePath: string, model: any): string` return a string from a parsed template
    * The function `ccode.TemplateToFile(templatePath: string, model: any, filePath: string, override: bool = true)` save a file from parsed template 
    * The function `ccode.StringToFile(input: string, filePath: string, override: bool = true)` save a file from a string

## Stage 2 - Add MCP server to the tool

## Stage 3 - Add SQL database model and accelerators
Details to be determined.

## Stage 4 (Not confirmed) - Add AI Agents with support to MCP and A2A
Details to be determined.

## Components and Architecture for stage 1

* A CLI which read a yaml config file using cobra+viper+pflag+gopkg.in/yaml.v3
* Use Valgo branch next-0.8 (https://github.com/cohesivestack/valgo/tree/next-0.8) for validate CLI entries
* Use https://github.com/dop251/goja package and esbuild to support executing the typescript functions
* Use https://github.com/NikolaLohinski/gonja as template parser
* Use https://github.com/pb33f/libopenapi as the OpenApi model parser
* Tests with testify package

## Config

**Config sample**

```yaml
ccode_path: ccode # The path where the structure of the project resides. This accepts relative paths. Relative paths are resolved from the config file directory. By default is `ccode`.
output_path: . # The root path where will be saved the produced artifacts. This is relative to the path where ccode command runs. By default is `.`
```

## CLI

```bash
ccode --config [config-path] --ccode-path [path] --output-path [output-path] run [process]
```

## Agent skill installation

This repository includes an installable Agent Skill package for the `skills` CLI:

```bash
npx skills add cohesivestack/ccode
```

To install only the Cohesive Code skill:

```bash
npx skills add cohesivestack/ccode --skill cohesive-code
```
