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

[Website and documentation](https://cohesivecode.dev)

Start with:

- [Getting Started](https://cohesivecode.dev/getting-started/)
- [Project Layout](https://cohesivecode.dev/using-ccode/project-layout/)
- [Processes](https://cohesivecode.dev/using-ccode/processes/)
- [Accelerators](https://cohesivecode.dev/using-ccode/accelerators/)
- [AI Skill Index](https://cohesivecode.dev/ai-skill-index/)

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

This repository includes one complete Cohesive Code Agent Skill installable with [`npx skills`](https://github.com/vercel-labs/skills):

```bash
npx skills add cohesivestack/ccode
```

The installed `cohesive-code` skill covers the full workflow: workspace setup and diagnosis, generator authoring, process execution, accelerator application, validation, and accelerator-state conflict resolution.

You can also select it explicitly:

```bash
npx skills add cohesivestack/ccode --skill cohesive-code
```

## Docs

- Start Here
  - [Getting Started](https://cohesivecode.dev/getting-started/)
  - [Philosophy](https://cohesivecode.dev/philosophy/)
  - [AI Skill Index](https://cohesivecode.dev/ai-skill-index/)
- Using Cohesive Code
  - [Project Layout](https://cohesivecode.dev/using-ccode/project-layout/)
  - [Processes](https://cohesivecode.dev/using-ccode/processes/)
  - [Templates](https://cohesivecode.dev/using-ccode/templates/)
  - [Generation](https://cohesivecode.dev/using-ccode/generation/)
  - [Accelerators](https://cohesivecode.dev/using-ccode/accelerators/)
  - [OpenAPI Workflows](https://cohesivecode.dev/using-ccode/openapi-workflows/)
  - [Wrapper & Versions](https://cohesivecode.dev/using-ccode/wrapper-and-versions/)
- Reference
  - [CLI](https://cohesivecode.dev/reference/cli/)
  - [Configuration](https://cohesivecode.dev/reference/configuration/)
  - [Runtime API](https://cohesivecode.dev/reference/runtime-api/)
  - [Accelerator States](https://cohesivecode.dev/reference/accelerator-states/)
- Cookbook
  - [Overview](https://cohesivecode.dev/cookbook/)
  - [Minimal Process](https://cohesivecode.dev/cookbook/minimal-process/)
  - [OpenAPI Docs](https://cohesivecode.dev/cookbook/openapi-docs/)
  - [Accelerated Artifact](https://cohesivecode.dev/cookbook/accelerated-artifact/)
- About
  - [License](https://cohesivecode.dev/about/license/)

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

Cohesive Code is developed and maintained by [Cohesive Stack LLC](https://cohesivestack.com) and released under the [MIT License](https://cohesivecode.dev/about/license/).
