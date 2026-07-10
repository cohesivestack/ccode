---
title: Getting Started
description: Install Cohesive Code, initialize a workspace, and run a first TypeScript process.
---

Cohesive Code is an AI-enabled code generation CLI. It runs TypeScript processes from a workspace, renders Gonja templates, parses JSON and OpenAPI documents, and tracks generated artifacts that need human or agent adjustment.

> **Experimental stage:** Cohesive Code is not recommended for use yet. The project is unstable, changing constantly, and may introduce breaking changes to interfaces, generated support files, accelerator state behavior, and CLI commands without notice.

## Install

Install the wrapper script to `~/.local/bin/ccode`:

```bash
curl -fsSL https://raw.githubusercontent.com/cohesivestack/ccode/master/installer/install.sh | bash
```

From a repository clone, you can run the installer directly:

```bash
bash installer/install.sh
```

The wrapper resolves the binary version for each workspace and caches release binaries under `~/.cache/ccode/releases`.

## Initialize a workspace

Create the project support files with an explicit version:

```bash
ccode init ccode --version v0.1.0
```

The default layout is:

```text
ccode.yaml
ccode/
  tsconfig.json
  .ccode/
    .gitignore
    accelerators/
    build/
    lib/
      context.ts
      openapi.ts
```

`ccode init` writes the configuration file when it is missing, refreshes generated support files under `.ccode/lib/`, preserves an existing `tsconfig.json`, clears `.ccode/build/`, and leaves accelerator state untouched.

## Create a process

Create a TypeScript process under `ccode_path`. The process must export a default function with one `Context` parameter.

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  const model = { name: "Cohesive Code" };

  const preview = ctx.renderTemplate("templates/greeting.tpl", model);
  ctx.println(preview);
  ctx.generate("templates/greeting.tpl", "generated/greeting.txt", model);
}
```

Create the template:

```jinja
Hello {{ data.name }}!
```

Recommended layout:

```text
ccode/
  hello/
    generate.ts
  templates/
    greeting.tpl
```

## Run it

Invoke the process without the `.ts` extension:

```bash
ccode run hello/generate
```

The runtime compiles the TypeScript process, executes it inside the Go runner, prints the rendered preview, and writes `generated/greeting.txt` under `output_path`.

## Next steps

- Read [Project Layout](/using-ccode/project-layout/) before building a real workspace.
- Read [Processes](/using-ccode/processes/) for the TypeScript process contract.
- Read [Accelerators](/using-ccode/accelerators/) before generating files that should be adjusted by a person or agent.
- Read [AI Skill Index](/ai-skill-index/) if you are preparing a skill or agent workflow.
