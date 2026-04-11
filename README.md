# Cohesive Code

Cohesive Code is an AI-enabled code generation CLI built around TypeScript processes, template rendering, and accelerators.

> **Status:** Alpha/Experimental stage. Not ready for production use yet.

## What It Does

* Executes TypeScript processes from a project workspace.
* Renders Jinja templates into output files.
* Parses JSON and OpenAPI documents from process code.
* Supports accelerators for predictable generation of artifacts that are expected to be adjusted by humans or agents.

## Core Concepts

* **Process**: A TypeScript file with a default export:
  * `export default function main(ctx: Context) { ... }`
* **Scope**: Logical output namespace used by accelerators.
  * Default scope is the process filename (without `.ts`).
  * You can change it at runtime with `ctx.setScope(...)`.
* **Accelerator**: A generated artifact tracked in state to avoid unsafe overwrites.
  * Output target: `<output_path>/<scope>/<artifact_id>`
  * State file: `<hidden_path>/state/accelerators.json`
  * Optional instructions markdown can be attached per artifact.

## Runtime API (TypeScript)

Import `Context` from `@ccode/context`:

```ts
import type { Context } from "@ccode/context";
```

Available methods:

* `println(message: string)`
* `setScope(scopeName: string)`
* `scope(): string`
* `renderTemplate(templatePath: string, data: any): string`
* `generate(templatePath: string, filePath: string, data: any): void`
* `accelerate(id: string, templatePath: string, data: any, instructionsPath?: string): void`
* `parseJSONFromBytes(jsonBytes: number[]): Record<string, any>`
* `parseJSONFromString(jsonString: string): Record<string, any>`
* `parseJSONFromFile(filePath: string): Record<string, any>`
* `parseOpenAPIFromBytes(specBytes: number[]): OpenAPIDocument`
* `parseOpenAPIFromString(spec: string): OpenAPIDocument`
* `parseOpenAPIFromFile(filePath: string): OpenAPIDocument`

## Minimal Process Example

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  const model = ctx.parseOpenAPIFromFile("specs/api.yaml");

  ctx.generate("templates/summary.tpl", "generated/summary.md", { model });

  ctx.accelerate(
    "handlers.ts",
    "templates/handlers.tpl",
    { model },
    "instructions/handlers.md",
  );
}
```

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
hidden_path: .ccode # Internal state/build folder. By default is `.ccode`
```

## CLI

### Main commands

```bash
ccode --config [config-path] init [path]
ccode --config [config-path] --ccode-path [path] --output-path [output-path] run [process]
```

### Accelerator inspection commands

```bash
ccode list accelerated [scopeId]
ccode list instructions
ccode get accelerated <scopeId>:<artifactId> [--instructions]
ccode get instruction <path>
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
