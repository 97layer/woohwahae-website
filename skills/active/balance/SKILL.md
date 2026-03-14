---
name: balance
description: >
  Use when the goal is token efficiency at the right depth for Layer OS work,
  balancing concise exploration with complete verification.
---

# balance

## Purpose

Use this skill when the goal is token efficiency at the right depth, not maximum compression.

The target is the Layer OS sweet spot:

- thin on low-risk work
- deeper on high-risk work
- concise in wording
- complete in verification

Layer OS defaults for that sweet spot:

- `session bootstrap` before broad reading
- `knowledge` / `next` before `handoff`
- packet before repo-wide search
- narrow `rg` before long file walks
- full handoff only on demonstrated context gaps
- optional MCP only after Go/CLI/daemon paths stop being enough
- local operator seat should stay near 3 active lanes (`layer-osd`, `8081/admin` when needed, one builder lane) with one optional manual shell
- if stale tmux or MCP/tooling processes pile up, trim them before assuming the product runtime is broken

## Modes

### `lean-fast`

Use for:

- one-file or two-file changes
- obvious bug fixes
- naming, wiring, and small CLI changes

Default behavior:

1. read only the entry file and directly related file
2. search narrowly with `rg`
3. implement immediately
4. run the smallest useful verification

### `lean-medium`

Use for:

- normal feature work
- multi-file changes with clear ownership
- runtime and CLI changes without contract risk

Default behavior:

1. read one governing doc if needed
2. inspect 3-6 relevant files
3. compare one or two likely approaches
4. implement
5. run focused verification plus `go test ./...` when the change touches runtime behavior

### `deep`

Use for:

- contract changes
- auth, security, deploy, release, rollback
- schema or persistence changes
- unclear bugs with multiple plausible causes

Default behavior:

1. read governing docs and affected contracts first
2. inspect all directly affected surfaces
3. state risks before editing
4. implement carefully
5. run broad verification

## Decision Rule

Pick the lightest mode that still safely covers the risk.

- low risk -> `lean-fast`
- medium risk -> `lean-medium`
- high risk or unclear risk -> `deep`

If new evidence raises risk during the task, escalate the mode instead of forcing the original budget.

## Output Rule

Keep user-facing responses short by default:

1. conclusion
2. evidence
3. next action

Do not spend tokens re-explaining repository rules that are already stable unless they directly affect the decision.
