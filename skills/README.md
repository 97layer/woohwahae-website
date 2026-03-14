# Skills

Layer OS treats skills as an execution surface, not as decoration.

## Related Docs

- External wrapper/operator boundary: `docs/operator.md`
- Verifier read-only audit lane: `docs/verifier.md`

## Layout

- `skills/active/`
  - short facade skills that define the default working grammar
- `skills/imported/`
  - external utilities kept behind the active layer

## Active Skills

- `orchestrate`
  - Codex-first execution flow for CLI-first work
- `verify`
  - verification, contract checks, and release checks
- `browser`
  - browser work via Playwright
- `capture`
  - screenshot and capture flow
- `sequence`
  - explicit step-by-step reasoning and option comparison without relying on an external Sequential Thinking tool
- `balance`
  - token-efficient sweet-spot execution using `lean-fast`, `lean-medium`, and `deep` modes by task risk

## Rule

- prefer CLI first
- use an active skill before reaching into imported skills
- use imported skills as deep references, not as the default entrypoint
- keep planning and runtime control inside Layer OS contracts
- keep skill names short, role-based, and easy to infer
