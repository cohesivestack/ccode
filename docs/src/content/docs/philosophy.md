---
title: Philosophy
description: The design principles behind Cohesive Code.
---

Cohesive Code is built around a simple distinction: generation is useful when it is repeatable, but many generated artifacts become valuable only after a person or an agent edits them. The CLI gives both cases a first-class workflow.

## Deterministic orchestration

A Cohesive Code process is ordinary TypeScript. It reads inputs, normalizes them into a rendering model, and calls runtime functions. This keeps domain decisions in code instead of hiding them inside large templates.

Templates are for presentation. They should format already prepared data, not walk complex schemas, infer names, or decide how the system should behave.

## Explicit boundaries

The project uses a clear workspace boundary:

- `ccode_path` contains process source, templates, specs, and generated support files.
- `output_path` is where generated artifacts are written.
- `hidden_path` stores internal state and build cache.
- `ccode.yaml` records the workspace version and default paths.

This matters for agents. A skill can inspect the config, know where process source belongs, know where outputs appear, and avoid editing generated runtime support files by accident.

## Generation and acceleration are different tools

Use `generate` when the output should be overwritten every time the process runs.

Use `accelerate` when the output is a proposal. Accelerated artifacts are tracked with a generated snapshot and optional instructions. If the target file has been changed since the last generated snapshot, Cohesive Code does not overwrite it.

That safe-write behavior is the core of the system. It lets code generation continue without erasing local judgement, hand edits, or agent refinements.

## Agent-friendly metadata

Accelerator inspection commands expose exactly the information an agent needs:

- unresolved artifacts
- current state
- optional instruction markdown
- decoded proposed content
- machine-readable JSON via `--for-agent`

List commands intentionally do not dump generated content or instruction bodies. Agents can first discover work, then request a focused bundle for one artifact.

## Cohesion over automation

Cohesive Code is not trying to make one prompt produce a finished codebase. It is designed to make generation processes legible, rerunnable, inspectable, and adjustable.

The philosophy is:

- keep transformation logic explicit
- keep rendering shallow
- keep paths predictable
- preserve adjusted work
- expose enough state for reliable AI workflows
