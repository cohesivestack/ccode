---
title: CLI
description: Command reference for Cohesive Code.
---

## Root options

```bash
ccode --config <config-path>
ccode --ccode-path <path>
ccode --output-path <path>
```

`--config` chooses a specific YAML config file. `--ccode-path` and `--output-path` override config values for that invocation.

## Initialize

```bash
ccode init [path] --version <version>
```

Initializes a workspace, writes or updates `ccode.yaml`, refreshes generated support files, and prepares the build cache.

## Run

```bash
ccode run <process>
```

Runs a `*.process.ts` TypeScript process under `ccode_path`. The process path is relative and may omit `.process.ts`.

## List runnable processes

```bash
ccode list processes
```

Recursively lists runnable `*.process.ts` files under `ccode_path` as process identifiers without the common suffix. Results are sorted and emitted as a JSON array. Dot-prefixed directories and the configured `hidden_path` are not inspected.

## Resolve an accelerated artifact

```bash
ccode resolve <artifactId>
```

Resolves a pending accelerated artifact without changing its output file. Use this when the generated proposal requires no edits and can be accepted as-is. A later run marks it pending again if the generated content or attached instructions change.

## List accelerated artifacts

```bash
ccode list accelerated
ccode list accelerated --include-resolved
ccode list accelerated --for-agent
```

Lists unresolved accelerated artifacts by default. Use `--include-resolved` to include resolved artifacts.

## Get an accelerated artifact

```bash
ccode get accelerated <artifactId>
ccode get accelerated <artifactId> --instructions
ccode get accelerated <artifactId> --instructions --for-agent
```

Without `--instructions`, the command returns metadata. With `--instructions`, it returns a composed resolvement bundle containing instruction markdown and decoded accelerated content.

## Get an instruction file

```bash
ccode get instruction <path>
ccode get instruction <path> --for-agent
```

Reads raw instruction markdown. The path is resolved using Cohesive Code's accelerator instruction rules.
