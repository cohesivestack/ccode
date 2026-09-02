---
name: cohesive-code
description: Work with Cohesive Code (`ccode`) workspaces end to end, including configuration, initialization, TypeScript processes, templates, OpenAPI and database generation, accelerators, CLI execution, output validation, troubleshooting, and accelerator-state conflict resolution. Use whenever a request involves Cohesive Code, the `ccode` CLI, `ccode.yaml`, or `*.accelerated.json` state.
---

# Cohesive Code

## Establish the workspace contract

1. Locate the applicable `ccode.yaml`; honor an explicit `--config` path.
2. Read the configuration before editing or running anything. Resolve `ccode_path`, `output_path`, `hidden_path`, and `version`, including overrides in the request.
3. Inspect existing inputs, targets, and Git status so project changes remain distinct from generated changes and pre-existing work.
4. Inspect the generated modules under `<ccode_path>/.ccode/lib/` when runtime
   API details matter. Use `context.ts` for `Context`; use the public module
   declarations for imports such as `@ccode/openapi`, `@ccode/database`,
   `@ccode/string`, `@ccode/go`, and `@ccode/typescript`. Do not import
   `internal/` modules.
5. Read `docs/src/content/docs/ai-skill-index.md`, then load only the pages it identifies for the task.

If local docs are unavailable, retrieve the matching Markdown from `https://raw.githubusercontent.com/cohesivestack/ccode/master/docs/src/content/docs/`. Prefer sources in this order: generated local types, checked-out implementation and CLI output, local docs, then older examples. Use the official `https://github.com/OAI/OpenAPI-Specification` reference only when local types and docs do not answer an OpenAPI schema question.

## Initialize, inspect, or diagnose

- For initialization and version selection, consult the getting-started, project-layout, configuration, and wrapper/version pages indexed by the local AI skill index.
- Verify uncertain CLI behavior with local `ccode <command> --help`, code, tests, or actual output.
- Treat the project as experimental and report documentation, type, CLI, or version mismatches rather than silently choosing stale behavior.
- Do not hand-edit `<ccode_path>/.ccode/lib/` or `<hidden_path>/build/`; refresh generated support files with `ccode init` when requested.
- Preserve accelerator state under `<hidden_path>/accelerators/` unless the task explicitly requires changing it.

## Author generation

Use this workflow for processes, templates, OpenAPI or database inputs, generated outputs, accelerators, scopes, and accelerator instructions.

- Keep process source, templates, specs, seed data, and instructions under `ccode_path`.
- Export the required default process function and import `Context` from `@ccode/context`.
- Parse each input once and normalize it into plain template data in TypeScript.
- Import `* as OpenAPI` from `@ccode/openapi` and use `OpenAPI.parseReference` when a process needs the document and fragment parts of a preserved `$ref`.
- Use `OpenAPI.Path.toColon`, `OpenAPI.Path.toSquareBrackets`,
  `OpenAPI.Path.toAngleBrackets`, or `OpenAPI.Path.toDollar` when generated
  framework paths need OpenAPI `{parameter}` syntax converted. Pass
  `{ omitLeadingSlash: true }` only when the target syntax omits the initial
  slash. These helpers do not interpolate values or validate parameter
  declarations.
- Import `* as Strings` from `@ccode/string` for general case and whitespace
  transformations. Import `* as Go` from `@ccode/go` for valid exported,
  unexported, or package names; do not recreate these transformations locally.
- Import `* as TypeScript` from `@ccode/typescript` and use
  `TypeScript.toTypeIdentifier` for PascalCase type, interface, class, enum,
  namespace, and type-alias names. Use `TypeScript.toValueIdentifier` for
  camelCase variable, function, method, parameter, local-constant, and
  intentionally normalized property names. These helpers sanitize declaration
  identifiers and protect reserved bindings; do not recreate that logic in a
  process or template. Control exports and visibility with TypeScript syntax,
  not identifier capitalization.
- Pass an initialism list to `Strings.camelCase`, `Strings.pascalCase`,
  `Strings.titleCase`, or `Strings.sentenceCase` when exact spellings such as
  `OpenAPI` or `GraphQL` must be preserved. Generic string helpers have no
  built-in initialisms. Go identifier helpers include the conventional Go
  initialisms; custom values extend that set and can override a default
  spelling. TypeScript identifier helpers have no built-in initialisms; pass
  custom values to preserve spellings such as `API`, `ID`, or `GraphQL`.
- Keep templates focused on presentation; move schema traversal, fallback naming, and branching decisions into TypeScript.
- For presentation-level naming in Gonja, use the built-in Cohesive Code filters
  (`camelCase`, `pascalCase`, `snakeCase`, `kebabCase`, `constantCase`,
  `dotCase`, `pathCase`, `titleCase`, `sentenceCase`, `upperFirst`,
  `lowerFirst`, `normalizeSpace`, `goExported`, `goUnexported`, and
  `goPackage`, `typeScriptType`, and `typeScriptValue`, plus `openAPIPathToColon`,
  `openAPIPathToSquareBrackets`, `openAPIPathToAngleBrackets`, and
  `openAPIPathToDollar`). Only the case filters and Go or TypeScript identifier
  filters documented as initialism-aware accept `initialisms=[...]`. Use the
  TypeScript-specific filters for declaration identifiers instead of rebuilding
  sanitization in templates. OpenAPI path filters accept the keyword-only
  boolean `omitLeadingSlash`.
- Use `ctx.generate(...)` only for files the generator may overwrite.
- Use `ctx.accelerate(...)` for proposals that must preserve later human or agent edits.
- Use stable artifact IDs. Call `ctx.setScope(...)` when the process filename is not a durable state namespace.
- Attach instructions only when downstream adjustment is required. State the target edits, constraints, preserved behavior, and verification steps.
- Keep instruction paths inside `ccode_path` and split large processes into focused helpers instead of growing template logic.

After changing generation inputs, run the exact process when safe, inspect every changed target, inspect unresolved accelerator state, and run the workspace's relevant checks.

## Run generation and apply accelerator work

1. Confirm the requested process path relative to `ccode_path`, then run it:

   ```bash
   ccode run <process>
   ```

2. Review command output and every changed target under `output_path`.
3. List unresolved artifacts and instructions:

   ```bash
   ccode list accelerated --for-agent
   ccode list instructions --for-agent
   ```

4. Handle `pending` items; skip `adjusted` unless the request includes resolved items. Diagnose `corrupt`, `ambiguous`, `missing_artifact`, or `missing_instructions` before editing.
5. Fetch each required bundle:

   ```bash
   ccode get accelerated <scopeId>:<artifactId> --instructions --for-agent
   ```

6. Treat `composed_markdown` as the primary bundle when present. Apply it to `output_path/<artifactId>` while preserving compatible human edits and repository conventions.
7. If the generated proposal requires no edits, accept it explicitly:

   ```bash
   ccode adjust <scopeId>:<artifactId>
   ```

8. Run checks named by the bundle and target project, then rerun both list commands and report remaining state.

Do not hand-edit accelerator state during normal generation work or force `pending: false`; use `ccode adjust` for that transition. Use `--include-resolved` only for a requested complete inventory. Stop before guessing when instructions are missing, state is ambiguous, or proposed content conflicts with intentional local edits.

## Resolve accelerator-state Git conflicts

Use this workflow only for merge, rebase, cherry-pick, or add/delete conflicts involving `<hidden_path>/accelerators/**/*.accelerated.json`.

1. Confirm the active Git operation and separate conflicts into generator inputs, output artifacts, and accelerator state. Keep unrelated conflicts out of scope.
2. Do not run a generator while its process, templates, specs, inputs, or instructions are conflicted.
3. For each state file, treat every non-empty line as a candidate for the scope and artifact encoded by its path.
4. Remove conflict markers and blank lines, deduplicate exact candidate lines, and retain multiple distinct valid JSON candidates when none is clearly obsolete. The CLI can report this as `ambiguous` for later repair.
5. Preserve every candidate as a complete record. Never combine `code`, `accelerated_checksum`, `instructions_checksum`, `instructions`, or `pending` values from different candidates.
6. Prefer retaining an added state file in an unclear add/delete conflict. Delete state only when the merged generator intentionally stops emitting the artifact or the user confirms removal. Do not infer obsolescence from a currently conflicted or renamed instruction file.
7. Validate each retained line as a complete JSON object and confirm that the path identifies the intended scope and artifact.
8. Once related conflicts are resolved, inspect state with the two list commands above. After the Git operation completes and inputs are coherent, rerun the affected process to rebuild or disambiguate state.

Report retained ambiguity and missing artifacts or instructions instead of hiding them.

## Always preserve

- Never edit `<ccode_path>/.ccode/lib/` or `<hidden_path>/build/` by hand.
- Preserve unrelated files, pre-existing changes, and compatible human edits.
- Keep recoverable accelerator candidates until generator inputs and the Git operation are coherent.
- Report mismatches and unresolved states explicitly.
