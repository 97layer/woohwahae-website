# Prompt Operating Model

## Intent

Layer OS does not treat prompts as ad-hoc thread prose.
It treats prompting as an operating surface with explicit owners, versioning expectations, evaluation loops, and input/output constraints.

This document defines the baseline rules for prompt work inside the Go-native Layer OS runtime.

## Related Docs

- Session baton semantics: `docs/legacy_session.md`
- External wrapper/operator boundary: `docs/operator.md`
- Verifier read-only boundary: `docs/verifier.md`
- Reusable worker prompt surface: `docs/agent-quickstart.md`
- Korean role-partition starter seeds: `docs/agent-role-seeds.md`
- Agent runtime contract surface: `docs/agent-integration.md`
- Validation/evidence surface: `docs/validation.md`

## Current Mapping

- `AGENTS.md`, `constitution/`, and stable operating rules carry long-lived behavior and tone constraints.
- runtime contracts and surfaces carry executable state and machine-visible context.
- the typed `prompting` block inside `SessionBootstrapPacket` and `AgentRunPacket` is now the runtime-owned behavior surface for active lanes.
- `review-room` tracks prompt-related decisions as agenda, not as prompt storage.
- `proposal` and `agent job` surfaces carry task-specific prompt intent when a lane becomes executable.

## Prompt Layers

### 1. Stable Instruction Layer

Use stable instruction surfaces for behavior that should survive across tasks:

- `AGENTS.md`
- constitution documents
- architecture and grammar rules
- reusable agent directives in `docs/`
- runtime contracts that define allowed state

Do not restate these in every task unless the task needs an exception.

### 2. Task Layer

Use task-local prompt content for:

- the exact outcome required now
- concrete constraints for the current lane
- minimal examples or target formats
- current runtime context that is not already encoded in contracts or state

Task-local prompt content belongs in:

- founder request text
- proposal summaries and notes
- agent job summaries, payloads, and result fields
- session bootstrap / job packet outputs

### 3. Evidence Layer

Prompt changes are not considered complete unless they have at least one of:

- a regression test
- a runtime audit expectation
- a review-room decision with evidence
- a proposal or job packet that shows how the prompt surface is meant to be consumed

## Operating Rules

### Centralize Prompt Definitions

When a prompt becomes reusable, it must stop living only in chat history.

Preferred order:

1. existing stable rule surface (`AGENTS.md`, constitution, docs)
2. runtime state or typed contract if the prompt is executable context
3. versioned proposal / agent job metadata if the prompt is lane-specific

Do not create stray root markdown files for prompt drafts.
If a prompt-related draft matters, absorb it into `review-room`, `proposal`, `agent job`, or `docs/`.

### Version Prompt Surfaces Intentionally

If a prompt is reused across lanes or time, give it an explicit identifier and revision trail.

Accepted mechanisms inside Layer OS:

- review-room item ids for decisions
- proposal ids for planning-stage prompt changes
- agent job ids for executable prompt lanes
- document history in git-backed docs for stable operating prompts

If external OpenAI prompt objects are adopted later, they should map back to the same local ids or notes.

### Keep Static Prefixes Stable

Repeated instructions and examples should stay structurally stable and come before variable user-specific content.
This keeps prompt caching potential high and prevents every request from becoming a fresh prompt shape.

In Layer OS terms:

- stable rules first
- task payload next
- volatile founder/session specifics last

### Keep Final Reports Founder-Readable

Final task reports should default to founder-readable language, not internal runtime shorthand.

- The first paragraph should explain what changed, why it matters, and what still needs attention.
- The first paragraph should also orient the founder: where the lane stands now, what matters next, and the recommended next move.
- Internal jargon, ids, file paths, route names, and command strings belong in evidence or supporting lines, not in the opening sentence.
- If a task still requires fixed result keys such as `summary` or `open_risks`, keep the keys and simplify the values.
- Role specialization does not excuse unreadable reporting; planner, implementer, verifier, and designer outputs should all stay understandable to a non-developer.

### Default To Practical Staff Posture

Reusable agent prompts should default to a practical staff aide / control-tower posture.

- sound like a partner helping run the business, not a raw terminal feed
- abandon the "passive phrase-router" fallback: actively parse intent from natural language and map it to runtime contracts instead of waiting for strict exact commands
- reduce system language unless it is the most precise evidence
- keep product, operations, infrastructure, brand, and user impact in view when deciding what matters
- make bounded proactive proposals when the next useful move is clear
- keep those proposals small and explicit instead of silently widening execution scope

### Default To State-First Reporting

When the founder asks for status, tracking, or what changed, the agent must answer as an operator first.

- lead with current system facts, changed files, runtime state, risks, and next watchpoints
- do not answer a status/tracking request with an implementation plan or redesign pitch
- correct scope drift immediately when the founder broadens the request beyond a local seam
- if the system state is unclear, say what is unknown and what evidence path is being checked

### Keep Internal Jargon Out Of The Opening

Internal design language should stay out of founder-facing openings unless it is the only precise term.

- avoid opening with phrases like `intent layer`, `router`, `canonical action`, or similar implementation shorthand
- if those terms are needed for evidence, put them after the plain-language situation summary
- do not let implementation self-talk replace the control-tower report

### Evaluate Prompt Changes

Prompt edits are production changes.
They require an eval loop proportional to their risk.

Minimum expectation:

- docs-only prompt rule -> review + audit pass
- CLI/API prompt surface change -> tests + surface audit
- behavior change affecting agent output -> targeted regression + review-room evidence
- reusable prompt workflow -> explicit eval plan or grader plan before broad rollout

### Constrain IO by Role

Prompt surfaces must stay role-specific.

- `planner`: broad context, synthesis, options, sequencing
- `implementer`: exact scope, file/module boundary, acceptance criteria
- `designer`: surface intent, experience goal, presentation constraints
- `verifier`: read-only or evidence-first framing, drift detection, failure classification

Do not hand the same giant undifferentiated prompt to every role.
The role boundary is part of the prompt contract.

At runtime, role metadata does not replace packet posture.
`role` remains useful for routing, token budget, and risk defaults, but active execution should follow the packet's typed `prompting` block whenever it is present.

### Keep Review Room as Decision Queue

`review-room` is not a long-form prompt memory store.
Use it to record:

- the existence of a prompt decision
- whether it is open / accepted / deferred / resolved
- the reason, rule, and evidence

Long prompt content or nuanced execution context should live in docs, proposals, or job packets instead.

## Implementation Guidance

### For New Reusable Prompt Work

Before adding a new prompt-heavy surface, check in order:

1. Can this be expressed as an update to existing rules?
2. Should this be a proposal instead of a freeform note?
3. Should this be an `AgentJob` lane with a `job packet` entrypoint?
4. Does it need a dedicated contract or only a typed payload field?

### For OpenAI-Facing Integrations

If Layer OS later adopts OpenAI prompt objects directly:

- keep a local prompt identifier in proposal/job/review notes
- record the external prompt id and version as evidence, not as the only source of truth
- rerun evals whenever the prompt version changes
- keep role-specific constraints in local operating docs even if the external prompt object stores the final text

## Acceptance Criteria for Prompt Work

A prompt-related change is considered structurally healthy when:

- it has an obvious home
- it does not create a new side-channel memory file
- its owner is clear
- its eval or audit path is clear
- its role boundary is clear
- its baton-pass path is clear
- a non-developer can understand the result without decoding internal system vocabulary first

## Immediate Gaps

The current Layer OS baseline is healthy but not complete.
The next gaps to close are:

- explicit prompt version registry for reusable prompt assets
- explicit eval/grader loop for prompt changes with behavioral risk
- stronger role-specific input/output budgets

Those gaps should be addressed as follow-on agenda, not by reintroducing freeform prompt sprawl.
