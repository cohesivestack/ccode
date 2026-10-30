# Cohesive Code Skills

This repository provides three complementary Agent Skills:

- [`ccode-author`](ccode-author/SKILL.md): initialize projects and create or modify processes, templates, inputs, and accelerator guides.
- [`ccode-produce`](ccode-produce/SKILL.md): run existing processes, then hand off pending artifacts to `ccode-resolve`.
- [`ccode-resolve`](ccode-resolve/SKILL.md): apply guides to existing pending artifacts, validate, and resolve them without running generation.

Author inputs with `ccode-author`. When execution is needed, it uses
`ccode-produce`, which runs the selected process and then invokes `ccode-resolve`.
Use `ccode-resolve` directly when generation has already completed. Its helper
only inspects existing state, so refreshing pending work never reruns a process.

## Install

```bash
npx skills add cohesivestack/ccode
```

Select individual skills explicitly. Install both `ccode-produce` and
`ccode-resolve` for generation followed by resolution:

```bash
npx skills add cohesivestack/ccode --skill ccode-author
npx skills add cohesivestack/ccode --skill ccode-produce
npx skills add cohesivestack/ccode --skill ccode-resolve
```

To preview the repository's installable skills without installing them:

```bash
npx skills add cohesivestack/ccode --list
```
