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

## Installation

### Wrapper installer (recommended)

Install the wrapper to `~/.local/bin/ccode` with `curl`:

```bash
curl -fsSL https://raw.githubusercontent.com/cohesivestack/ccode/master/installer/install.sh | bash
```

Or run the installer from this repository clone:

```bash
bash installer/install.sh
```

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

## Releases

Releases are automated with GoReleaser through GitHub Actions.

* Trigger: push a Git tag matching `v*`
* Accepted format: `vMAJOR.MINOR.PATCH` (for example, `v1.2.3`) and prereleases (for example, `v1.3.0-rc1`)
* Published assets:
  * `linux/amd64`
  * `linux/arm64`
  * `darwin/amd64`
  * `darwin/arm64`
  * `windows/amd64`
* Each release includes platform archives and `checksums.txt` (SHA-256)

**Create a release**

```bash
git tag v1.2.3
git push origin v1.2.3
```

### How the wrapper works

The wrapper at `installer/bin/ccode`:

* Resolves version precedence in this order:
  * `CCODE_VERSION`
  * nearest `.ccode/version` (searching upward from the current directory)
  * `~/.config/ccode/version`
  * highest cached version in `~/.cache/ccode/releases`
  * latest stable GitHub release (non-draft, non-prerelease)
* Caches binaries under `~/.cache/ccode/releases`
* For normal execution, does not modify `.ccode/version` or `~/.config/ccode/version`
* Supports pinning:
  * `ccode pin`
  * `ccode pin <version>`
  * `ccode pin latest`
  * `ccode pin --global`
  * `ccode pin <version> --global`

## Install from release assets

1. Open the GitHub Releases page: `https://github.com/cohesivestack/ccode/releases`
2. Download the archive for your platform
3. Verify the downloaded file with `checksums.txt`
4. Extract and place the `ccode` binary in your `PATH`

## Agent skill installation

This repository includes an installable Agent Skill package for the `skills` CLI:

```bash
npx skills add cohesivestack/ccode
```

To install only the Cohesive Code skill:

```bash
npx skills add cohesivestack/ccode --skill cohesive-code
```
