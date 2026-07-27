---
name: author-ccode-generation
description: Author or revise Cohesive Code generator inputs, including TypeScript processes, Gonja templates, OpenAPI inputs, `ctx.generate` outputs, `ctx.accelerate` artifacts, scopes, and accelerator instructions. Use when designing or changing the generator side of a `ccode` workspace before or alongside validating a process.
---

# Author Cohesive Code Generation

## Prepare

1. Read `ccode.yaml` and resolve `ccode_path`, `output_path`, and `hidden_path`.
2. Inspect the existing process, templates, inputs, instructions, and target artifacts before editing.
3. Inspect `<ccode_path>/.ccode/lib/context.ts` to confirm the installed runtime contract.
4. Read `docs/src/content/docs/ai-skill-index.md`, then load only its authoring pages relevant to the request.
5. If local docs are unavailable, retrieve the matching file from `https://raw.githubusercontent.com/cohesivestack/ccode/master/docs/src/content/docs/`.

Use the official `https://github.com/OAI/OpenAPI-Specification` reference only when local types and docs do not answer an OpenAPI field or schema question.

## Implement

- Keep process source, templates, specs, seed data, and instructions under `ccode_path`.
- Export the required default process function and import `Context` from `@ccode/context`.
- Parse each input once and normalize it into plain template data in TypeScript.
- Keep templates focused on presentation; move schema traversal, fallback naming, and branching decisions into TypeScript.
- Use `ctx.generate(...)` only for files the generator may overwrite.
- Use `ctx.accelerate(...)` for proposals that must preserve later human or agent edits.
- Use stable artifact IDs. Call `ctx.setScope(...)` when the process filename is not a durable state namespace.
- Attach instructions only when downstream adjustment is required. State the target edits, constraints, preserved behavior, and verification steps.
- Keep instruction paths inside `ccode_path`.
- Split large processes into focused helpers instead of growing template logic.
- Never edit `<ccode_path>/.ccode/lib/` or `<hidden_path>/build/` by hand.

## Validate

1. Run the exact process when doing so is safe:

   ```bash
   ccode run <process>
   ```

2. Inspect every generated or accelerated target changed by the process.
3. Inspect unresolved accelerator state:

   ```bash
   ccode list accelerated --for-agent
   ccode list instructions --for-agent
   ```

4. Run the workspace's relevant type checks or tests.
5. Report any mismatch among generated types, CLI behavior, docs, and examples.
