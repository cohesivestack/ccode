---
title: Wrapper & Versions
description: How the ccode wrapper chooses and caches CLI versions.
---

The installed `ccode` command is a wrapper. Its job is to choose the right release binary for the current workspace and execute it.

## Version precedence

The wrapper resolves versions in this order:

1. `--version` flag
2. nearest `ccode.yaml` `version` entry, or the file passed with `--config`
3. global pin at `~/.config/ccode/version`
4. latest stable GitHub release

For normal execution, the wrapper does not modify `ccode.yaml` or the global version file.

## Workspace version

`ccode.yaml` must include a version:

```yaml
version: v0.1.0
```

`ccode init` writes or updates the version entry and refreshes generated support files for that version.

## Pinning

The wrapper supports:

```bash
ccode pin
ccode pin v0.1.0
ccode pin latest
```

`ccode pin` writes the global version file at `~/.config/ccode/version`.

## Cache

Release binaries are cached under:

```text
~/.cache/ccode/releases
```

This lets different workspaces use different versions without requiring a separate manual installation per project.
