# Cohesive Code Skills

This repository exposes installable Agent Skills for `npx skills`.

## Install

```bash
# from GitHub shorthand
npx skills add cohesivestack/ccode

# install only this skill
npx skills add cohesivestack/ccode --skill cohesive-code

# list available skills without installing
npx skills add cohesivestack/ccode --list
```

## Skill layout

Each installable skill is self-contained under `skills/<skill-name>/` and includes:

- `SKILL.md`
- optional `scripts/`
- optional `references/`
- optional `examples/`
- optional `agents/`
