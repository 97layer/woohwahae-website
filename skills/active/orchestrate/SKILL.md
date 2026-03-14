---
name: orchestrate
description: >
  Use when Codex is driving a non-trivial implementation lane inside Layer OS
  and needs the default local orchestration workflow.
---

# orchestrate

## Purpose

Use this skill when Codex is driving a non-trivial implementation lane inside Layer OS.

## Defaults

- CLI first
- contracts before code
- active constitution before imported detail
- verify before cutover

## Workflow

1. read `constitution/charter.md`
2. read `docs/architecture.md`
3. read `skills/manifest.json`
4. create an internal preflight record before structural work
5. use terminal commands before inventing new tooling

## Default Command

```bash
go run ./cmd/layer-osctl preflight create --id <id> --task "<task>" --steps load-contracts,run-tests --risks <risk> --checks external-verify
```

## Rule

Codex is the orchestrator.
Imported skills are external utilities only.
