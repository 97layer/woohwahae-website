# Expansion

Historical archive only.
This note is kept for continuity and should not be treated as active runtime authority or default read-path material.

## Intent

This document fixes the next expansion boundary for Layer OS without importing old-system residue.

## Related Docs

- Legacy session port: `docs/legacy_session.md`
- External wrapper/operator guardrails: `docs/operator.md`
- Verifier read-only audit lane: `docs/verifier.md`

Kernel first.
Surfaces second.
AI infrastructure third.

## Surface Rule

Future surfaces do not own state.
They consume Layer OS contracts.

### Planned Surfaces

1. `website`
   - public read surface
   - consumes release artifacts and selected public state
2. `cockpit`
   - founder and operator surface
   - same-binary founder console in the first wave
   - consumes auth, preflight, work, approval, policy, execute, release, deploy, and handoff
3. `app`
   - narrow founder/mobile surface
   - consumes the same kernel contracts as cockpit
4. `public api`
   - external integration surface
   - only after contracts stabilize

## AI Infrastructure Rule

AI infrastructure is not a separate kingdom.
It is a set of lanes behind the kernel.

### Planned Lanes

1. `preflight`
   - planning and risk record before work starts
2. `memory`
   - current focus, next steps, open risks, handoff
3. `verify`
   - tests, checks, critique, recorded verification runs
4. `policy`
   - deterministic routing and escalation records
5. `gateway`
   - recorded provider and model call intent behind `single`/`go` policy decisions; no real fan-out adapter yet
6. `execute`
   - contract-aware work execution records

## Migration Rule

- old-system paths may appear only in temporary bridge skills
- bridge skills must live outside `internal/`
- once a kernel-native contract exists, the old bridge must be removed

## Current Priority

1. internalize execute as a kernel-native lane
2. add same-binary cockpit for the founder loop
3. keep app/public planning contract-first
4. add live providers only after kernel records stay stable
