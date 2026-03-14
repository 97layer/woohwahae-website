# Agent Integration Standard

## Contract

For the reusable worker prompt surface used by both explicit-task and self-start agents, see `docs/agent-quickstart.md`.
For Korean role-partition starter prompts, see `docs/agent-role-seeds.md`.
For older underscore-path references, see `docs/agent_integration.md`.

An external agent (Python or otherwise) touches three contracts and nothing else.

1. **Receives**: `AgentRunPacket` via `layer-osctl job packet`
2. **Reports**: `POST /api/layer-os/jobs/report` with `AgentJob.result` populated
3. **Fails**: reports `status=failed`; the runtime promotes the job to review-room automatically

## AgentRunPacket Field Usage

| Field             | When to use                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| `job`             | Scope, role, payload — the task definition                                  |
| `runtime`         | Report path, report command, terminal statuses, auth hint                   |
| `knowledge`       | Low-token current focus, risks, action routes, and recent context           |
| `prompting`       | Typed execution posture. Treat this as the runtime-owned behavior contract. |
| `handoff`         | Full baton. Present when the lane truly needs wider continuity context.     |
| `handoff_summary` | Slim baton for non-planner lanes: focus, risks, top review items, counts    |
| `proposal`        | When the job was promoted from a proposal — source ref                      |

`runtime.report_path` and `runtime.report_command` are the authoritative targets.
Do not hardcode `/api/layer-os/jobs/report` — read it from the packet.
In Antigravity-like sandboxes where direct `.layer-os/*` reads may fail, keep the baton on localhost: pull `layer-osctl job packet` (or the packet URL) and report back through `runtime.report_path` / `runtime.report_command` instead of reading runtime files directly.
When using `layer-osctl job report`, prefer `--result-file` or `--result-json` so arrays/objects in the official result contract survive intact.
The report response now includes `follow_up`, a runtime-owned hint that tells the agent or operator what to look at next after the terminal report lands.

`prompting` is now the packet-first behavior surface.
Use it before role seeds or quickstart prose when the packet is available.
`role` still matters for routing, risk, and token budgets, but the packet's `prompting` block is the authoritative execution posture for that lane.
Session bootstrap defaults stay conservative (`single_step`, read-only), while executable job packets default to `multi_step` inside the current lane unless the packet narrows that budget explicitly.

An agent must not write back to runtime state directly.
All output flows through the report endpoint.

## AgentJob.result Contract

`result` is an untyped object at the transport layer, but the canonical worker/report contract is fixed.
Quickstart prompts, role seeds, operators, and external agents should all emit the same seven-key shape below.

### Official Extended Contract

| Key             | Type            | Description                                                           |
|-----------------|-----------------|-----------------------------------------------------------------------|
| `summary`       | string          | One-sentence completion or failure statement in plain language first |
| `artifacts`     | array           | Produced outputs — file paths, IDs, or refs (empty if none)         |
| `verification`  | array or object | Checks run plus a plain-language explanation of what the outcome means |
| `open_risks`    | array           | Remaining risks, unknowns, or follow-up cautions described in impact terms first |
| `follow_on`     | array           | Exactly one bounded next improvement candidate stated in plain language, or empty if none |
| `touched_paths` | array           | Files or paths intentionally changed for this job                    |
| `blocked_paths` | array           | Paths intentionally not changed because they were blocked or out-of-scope |

### Contract Notes

- `summary` and `artifacts` remain the minimal runtime-safe floor for generic transport compatibility.
- The official Layer OS worker contract is the extended seven-key shape above; prompts and human wrappers should not invent alternate result keys.
- Result keys stay fixed, but the values should be readable by a non-developer founder or operator without opening the code first.
- Agents should sound like a practical staff aide and control tower: explain the situation, the implication, and the best next move before dumping system evidence.
- Put internal ids, route names, command strings, and file paths after the plain-language explanation, not instead of it.
- Runtime normalization may still accept legacy evidence aliases such as `changed_paths` for compatibility, but agents should emit `touched_paths` and `blocked_paths` as the canonical keys.
- Runtime-owned dispatch metadata may later enrich the stored result with provenance such as `provider`, `gateway_call_id`, `completion_mode`, or `council`; worker-authored result payloads should still start from the canonical seven-key shape above.
- If a field has no entries, send an empty array rather than omitting the key.

Status transitions an agent may send (`runtime.terminal_statuses`):

- `succeeded` — work complete, result populated
- `failed` — work could not be completed, summary explains why in plain language
- `canceled` — work abandoned before completion; summary explains why in plain language

Jobs that never report surface as risk in handoff and knowledge packets.

## LLM Loop Ownership

Python agents own their own LLM loop.
The Go runtime does not record LLM calls made inside a Python agent.
The Go runtime records one thing at the dispatch boundary: a `GatewayCall` that marks
the intent to hand the packet to the agent.

See `GatewayCall Role` in `docs/architecture.md`.

## Failure Path

1. Agent reports `status=failed` → runtime auto-creates a review-room item
2. Agent crashes without reporting → founder manually transitions the job to `failed`
3. In either case the job must close as `done`, `failed`, or `needs_review`

## Role-Specific Payload Guidance

| Role          | `job.payload` should contain                           |
|---------------|--------------------------------------------------------|
| `planner`     | Scope statement, referenced `proposal_id`, option list, optional `chain_rules` |
| `implementer` | Exact file/module boundary, acceptance criteria        |
| `verifier`    | Target `job_id`, expected outcome, allowed drift       |
| `designer`    | Brand/surface goal, target experience, review rubric   |

Prompt structure inside the agent follows the same layer order as `docs/prompting.md`:
stable rules first, packet `prompting` plus task payload next, volatile session context last.

## Optional Council Payload

For a bounded single-thread multi-model pass, `job.payload` may include an optional `council` object.

```json
{
  "council": {
    "providers": ["claude", "openai"],
    "primary_provider": "claude"
  }
}
```

- The runtime dispatches the packet sequentially to each listed provider.
- `primary_provider` selects which provider's response becomes the top-level job result and gateway return.
- The runtime also stores a compact `result.council` summary so operators can inspect the full council trail without re-opening gateway logs.
- This is a runtime-owned dispatch extension, not a worker-side reporting channel.
- When the payload omits `council`, daemon dispatch may still derive the same bounded council automatically for non-`founder.manual` `planner` / `designer` lanes when multiple direct providers are dispatch-ready. In that case the stored result also includes `council_auto=true`.
