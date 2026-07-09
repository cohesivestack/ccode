---
title: CLI
description: Command reference for Cohesive Code.
slug: 0.1/reference/cli
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

Runs a TypeScript process under `ccode_path`. The process path is relative and may omit `.ts`.

## List accelerated artifacts

```bash
ccode list accelerated [scopeId]
ccode list accelerated [scopeId] --include-resolved
ccode list accelerated [scopeId] --for-agent
```

Lists unresolved accelerated artifacts by default. Use `--include-resolved` to include adjusted artifacts.

## List instruction references

```bash
ccode list instructions
ccode list instructions --include-resolved
ccode list instructions --for-agent
```

Lists instruction references attached to accelerated artifacts.

## Get an accelerated artifact

```bash
ccode get accelerated <scopeId>:<artifactId>
ccode get accelerated <scopeId>:<artifactId> --instructions
ccode get accelerated <scopeId>:<artifactId> --instructions --for-agent
```

Without `--instructions`, the command returns metadata. With `--instructions`, it returns a composed adjustment bundle containing instruction markdown and decoded accelerated content.

## Get an instruction file

```bash
ccode get instruction <path>
ccode get instruction <path> --for-agent
```

Reads raw instruction markdown. The path is resolved using Cohesive Code's accelerator instruction rules.
