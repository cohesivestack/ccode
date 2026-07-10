# Cohesive Code

Cohesive Code is an AI-enabled code generation CLI built around TypeScript processes, Gonja templates, OpenAPI parsing, and accelerators for human or agent adjustment.

It is designed for deterministic generation where generator authors control the process code and templates, while downstream developers or agents can safely inspect and adjust generated proposals.

> **Experimental stage:** Cohesive Code is not recommended for use yet. The project is unstable, changing constantly, and may introduce breaking changes to interfaces, generated support files, accelerator state behavior, and CLI commands without notice.

## Quick example

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  const spec = ctx.parseOpenAPIFromFile("specs/api.yaml");

  ctx.generate("templates/summary.tpl", "generated/summary.md", { spec });

  ctx.accelerate(
    "handlers.ts",
    "templates/handlers.tpl",
    { spec },
    "instructions/handlers.md",
  );
}
```

Run the process and inspect accelerator work:

```bash
ccode run api/generate
ccode list accelerated --for-agent
ccode get accelerated generate-api:handlers.ts --instructions --for-agent
```

## Website and documentation

The documentation source lives in [docs/src/content/docs](docs/src/content/docs).

Start with:

- [Getting Started](docs/src/content/docs/getting-started.md)
- [Project Layout](docs/src/content/docs/using-ccode/project-layout.md)
- [Processes](docs/src/content/docs/using-ccode/processes.md)
- [Accelerators](docs/src/content/docs/using-ccode/accelerators.md)
- [AI Skill Index](docs/src/content/docs/ai-skill-index.md)

## Installing

Install the wrapper to `~/.local/bin/ccode`:

```bash
curl -fsSL https://raw.githubusercontent.com/cohesivestack/ccode/master/installer/install.sh | bash
```

From this repository clone:

```bash
bash installer/install.sh
```

The wrapper resolves the binary version for each workspace and caches release binaries under `~/.cache/ccode/releases`.

## Agent skills

This repository includes Cohesive Code Agent Skills installable with [`npx skills`](https://github.com/vercel-labs/skills):

```bash
npx skills add cohesivestack/ccode
```

Install only the broad workspace skill:

```bash
npx skills add cohesivestack/ccode --skill cohesive-code
```

Focused workflow skills:

```bash
npx skills add cohesivestack/ccode --skill author-ccode-generation
npx skills add cohesivestack/ccode --skill run-ccode-generation
npx skills add cohesivestack/ccode --skill merge-ccode-accelerator-state
```

## Docs

- Start Here
  - [Getting Started](docs/src/content/docs/getting-started.md)
  - [Philosophy](docs/src/content/docs/philosophy.md)
  - [AI Skill Index](docs/src/content/docs/ai-skill-index.md)
- Using Cohesive Code
  - [Project Layout](docs/src/content/docs/using-ccode/project-layout.md)
  - [Processes](docs/src/content/docs/using-ccode/processes.md)
  - [Templates](docs/src/content/docs/using-ccode/templates.md)
  - [Generation](docs/src/content/docs/using-ccode/generation.md)
  - [Accelerators](docs/src/content/docs/using-ccode/accelerators.md)
  - [OpenAPI Workflows](docs/src/content/docs/using-ccode/openapi-workflows.md)
  - [Wrapper & Versions](docs/src/content/docs/using-ccode/wrapper-and-versions.md)
- Reference
  - [CLI](docs/src/content/docs/reference/cli.md)
  - [Configuration](docs/src/content/docs/reference/configuration.md)
  - [Runtime API](docs/src/content/docs/reference/runtime-api.md)
  - [Accelerator States](docs/src/content/docs/reference/accelerator-states.md)
- Cookbook
  - [Overview](docs/src/content/docs/cookbook/index.mdx)
  - [Minimal Process](docs/src/content/docs/cookbook/minimal-process.md)
  - [OpenAPI Docs](docs/src/content/docs/cookbook/openapi-docs.md)
  - [Accelerated Artifact](docs/src/content/docs/cookbook/accelerated-artifact.md)
- About
  - [License](docs/src/content/docs/about/license.md)

## Releases

Releases are automated with GoReleaser through GitHub Actions.

- Trigger: push a Git tag matching `v*`
- Accepted format: `vMAJOR.MINOR.PATCH`, for example `v1.2.3`, and prereleases such as `v1.3.0-rc1`
- Published assets:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`

Each release includes platform archives and `checksums.txt`.

Create a release:

```bash
git tag v1.2.3
git push origin v1.2.3
```

## GitHub Code Contribution Guide

We welcome contributions to the project. To make review smooth, please follow these guidelines:

- **Discuss larger changes first**: Open an issue or discussion before starting broad design or behavior changes.
- **Make commits small and cohesive**: Keep each commit focused on one task or change.
- **Format Go code**: Run `gofmt` on changed Go files.
- **Cover behavior with tests**: Add or update tests for runtime, CLI, wrapper, or accelerator-state changes.
- **Update docs and skills when behavior changes**: Keep `docs/src/content/docs`, README, and installable skills aligned with user-facing behavior.
- **Use respectful language**: Keep collaboration direct, specific, and constructive.

## License

Copyright © 2026 Carlos Forero

Cohesive Code is developed and maintained by [Cohesive Stack LLC](https://cohesivestack.com) and released under the [MIT License](docs/src/content/docs/about/license.md).
