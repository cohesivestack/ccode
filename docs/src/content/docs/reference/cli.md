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

Runs a `*.proc.ts` TypeScript process under `ccode_path`. The process path is relative and may omit `.proc.ts`.

## List runnable processes

```bash
ccode proc list
```

Recursively lists runnable `*.proc.ts` files under `ccode_path` as process identifiers without the common suffix. Results are sorted and emitted as a JSON array. Dot-prefixed directories and the configured `hidden_path` are not inspected.

## Resolve an accelerated artifact

```bash
ccode acc resolve <artifactId>
```

Resolves a pending accelerated artifact without changing its output file. Use this when the generated proposal requires no edits and can be accepted as-is. A later run marks it pending again if the generated content or attached guide changes.

## List accelerated artifacts

```bash
ccode acc list
ccode acc list --include-resolved
ccode acc list --for-agent
```

Lists unresolved accelerated artifacts by default. Use `--include-resolved` to include resolved artifacts.

## Show an accelerated artifact

```bash
ccode acc show <artifactId>
ccode acc show <artifactId> --guide
ccode acc show <artifactId> --guide --for-agent
```

Without `--guide`, the command returns metadata. With `--guide`, it returns a composed resolvement bundle containing guide markdown and decoded accelerated content.

## Show an artifact’s guide

```bash
ccode acc guide <artifactId>
ccode acc guide <artifactId> --for-agent
```

Looks up the accelerated artifact by ID and reads its attached guide as raw Markdown. With `--for-agent`, returns JSON containing the guide’s `path` and `markdown`. Returns an error if the artifact does not exist, has no attached guide, or its guide file is missing.
