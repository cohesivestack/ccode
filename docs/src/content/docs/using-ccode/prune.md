---
title: Pruning
description: Remove stale output files while preserving artifacts emitted during the current run.
---

Call `ctx.prune(...)` after generating and accelerating artifacts to remove obsolete
outputs. Pruning matches files under the configured `output_path` and preserves
artifacts emitted so far during the current top-level process run, including
artifacts from processes called through `ctx.runProcess(...)`.

## API

```ts
ctx.prune(pathPatterns);
```

`pathPatterns` is an array of strings. Each string is a regular expression matched against the file's path relative to
`output_path`. Paths use `/` separators on every platform and have no leading
`./`. A file is selected if any pattern matches. Patterns use Go regular
expression syntax; they are strings, not JavaScript `RegExp` objects or globs.
Use `^` and `$` when a pattern should match the entire path.

## Example

Use `.gen.<extension>` for generated outputs and `.acc.<extension>` for
accelerated outputs so pruning patterns are easy to recognize:

```ts
import type { Context } from "@ccode/context";

export default function main(ctx: Context) {
  const model = ctx.parseJSONFromFile("data/model.json");

  ctx.generate("templates/model.tpl", "src/model.gen.go", model);
  ctx.accelerate("src/handler.acc.go", "templates/handler.tpl", model);
  ctx.runProcess("api/generate");

  ctx.prune(["\\.gen\\.go$", "\\.acc\\.go$"]);
}
```

The double backslashes escape regex dots inside TypeScript strings. These
patterns select `.gen.go` and `.acc.go` files anywhere under `output_path`.
To restrict pruning to `src/`, use
`["^src/.*\\.gen\\.go$", "^src/.*\\.acc\\.go$"]`.

For example, if `src/old-model.gen.go` and `src/old-handler.acc.go` exist but
neither is emitted in this run, pruning deletes them. It keeps
`src/model.gen.go`, `src/handler.acc.go`, and matching artifacts emitted by
`api/generate`. Files such as `src/helpers.go` do not match and stay in place.

## Behavior

- Successful `generate` calls protect their output paths. Successful `accelerate`
  calls protect their output paths and state files, including when an existing
  artifact is preserved because it was edited or already has the expected content.
- Tracking is shared by the parent and all nested processes. It resets for each
  new top-level run and is kept only in memory.
- Pruning runs immediately. Place it after all relevant generation and child
  process calls so their outputs are protected.
- Matching files that were not emitted in this run are deleted even if manually
  edited or never previously tracked by Cohesive Code. Deleting an accelerated
  artifact also removes its corresponding `.accelerated.json` state file.
- Directories and symbolic links are left in place. Pruning does not follow
  symbolic links inside the output directory. A configured `output_path` that
  is itself a symbolic link is resolved before scanning.
- An empty pattern array or missing output directory does nothing. Every pattern
  is validated before deletion; an invalid pattern raises an error. Filesystem
  errors also raise errors, and earlier deletions are not rolled back.

Refresh an existing workspace's TypeScript support files with `ccode init` after
upgrading to use `prune`.
