# Cohesive Code Skills

This repository provides two complementary Agent Skills:

- [`ccode-author`](ccode-author/SKILL.md): initialize projects and create or modify processes, templates, inputs, and accelerator guides.
- [`ccode-produce`](ccode-produce/SKILL.md): run existing processes, apply guided accelerated changes, validate, and resolve artifacts.

## Install

```bash
npx skills add cohesivestack/ccode
```

Select either skill explicitly:

```bash
npx skills add cohesivestack/ccode --skill ccode-author
npx skills add cohesivestack/ccode --skill ccode-produce
```

To preview the repository's installable skills without installing them:

```bash
npx skills add cohesivestack/ccode --list
```
