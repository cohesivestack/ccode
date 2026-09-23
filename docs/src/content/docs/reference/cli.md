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

Resolves a pending accelerated artifact without changing its output file. Use this when the generated proposal requires no edits and can be accepted as-is. A later run marks it pending again if the generated content or attached guide changes.

## List accelerated artifacts

```bash
ccode list acc
ccode list acc --include-resolved
ccode list acc --for-agent
```

Lists unresolved accelerated artifacts by default. Use `--include-resolved` to include resolved artifacts.

## Get an accelerated artifact

```bash
ccode get accelerated <artifactId>
ccode get accelerated <artifactId> --guide
ccode get accelerated <artifactId> --guide --for-agent
```

Without `--guide`, the command returns metadata. With `--guide`, it returns a composed resolvement bundle containing guide markdown and decoded accelerated content.

## Get a guide file

```bash
ccode get guide <path>
ccode get guide <path> --for-agent
```

Reads raw guide markdown. The path is resolved using Cohesive Code's accelerator guide rules.
