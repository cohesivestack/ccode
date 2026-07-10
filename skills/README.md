# Cohesive Code Skills

This repository exposes installable Agent Skills for `npx skills`.

## Install

```bash
# from GitHub shorthand
npx skills add cohesivestack/ccode

# install only the broad workspace skill
npx skills add cohesivestack/ccode --skill cohesive-code

# install a focused workflow skill
npx skills add cohesivestack/ccode --skill author-ccode-generation
npx skills add cohesivestack/ccode --skill run-ccode-generation
npx skills add cohesivestack/ccode --skill merge-ccode-accelerator-state

# list available skills without installing
npx skills add cohesivestack/ccode --list
```

## Included skills

- `cohesive-code`: broad workspace orientation and doc routing.
- `author-ccode-generation`: author TypeScript processes, templates, OpenAPI workflows, accelerators, and instructions.
- `run-ccode-generation`: run processes, inspect outputs, and apply accelerator instruction bundles.
- `merge-ccode-accelerator-state`: resolve Git conflicts in accelerator state files.

## Skill layout

Each installable skill is self-contained under `skills/<skill-name>/` and includes:

- `SKILL.md`
- optional `agents/`
