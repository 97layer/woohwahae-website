# Architecture

## Intent

Layer OS is not a port of `97layerOS`.
It is a derived rebuild.

The constitution is inherited.
The runtime is redesigned.

## Related Docs

- Legacy session port: `docs/legacy_session.md`
- External wrapper/operator boundary: `docs/operator.md`
- Verifier read-only audit lane: `docs/verifier.md`

## System Spine

The runtime is built around thirty canonical contracts:

1. `CompanyState`
2. `ProposalItem`
3. `WorkItem`
4. `ApprovalItem`
5. `ReleasePacket`
6. `DeployRun`
7. `DeployTarget`
8. `EventEnvelope`
9. `SystemMemory`
10. `AuthStatus`
11. `PreflightRecord`
12. `PolicyDecision`
13. `GatewayCall`
14. `ExecuteRun`
15. `VerificationRecord`
16. `RollbackRun`
17. `FlowRun`
18. `FounderView`
19. `FounderSummary`
    - includes `priority_rationale` so the active founder lane is evidence-backed, not implicit
20. `ReviewRoom`
    - review items may carry `rationale`, `resolution`, and `evidence` so both auto-ingested and manually closed agenda remains explainable
    - founder action events also carry explicit `reason`, `rule`, and `evidence` metadata so approval, release, and rollback decisions are evidence-backed in runtime history
21. `ReviewRoomSummary`
22. `AdapterSummary`
23. `CapabilityRegistry`
    - binds actors, declared providers, and adapters as runtime data so roster changes stay detachable
24. `KnowledgePacket`
    - compact session bootstrap packet derived from memory, founder priority, review-room, parallel lanes, and capabilities
25. `SessionBootstrapPacket`
26. `SnapshotPacket`
27. `WriteLease`
28. `HandoffPacket`
29. `DaemonStatus`
    - a process-local heartbeat contract for uptime/address/degraded-reason visibility and optional architect-loop telemetry; it is observable through API/CLI but is not part of snapshot continuity state
30. `ChainRules`
    - a per-job follow-up contract that lets `job.report` create the next queued agent lane without inventing a second planner queue

## Compact Scaffold

Layer OS keeps one compact full-stack harness instead of scattering authority.

- source scaffold: `constitution`, `contracts`, `docs`, `skills`, `cmd`, `internal`, `scripts`
- runtime scaffold: `internal/runtime` + `internal/api` + `cmd/layer-osctl` + `cmd/layer-osd`
- runtime data scaffold: `.layer-os/` only
- canonical surface catalog: `internal/runtime/surface_spec.go`

That means operator docs, CLI dispatch, API routes, and JSON contracts should all collapse back onto the same scaffold instead of becoming parallel encyclopedias.

## Contract Tiers

The thirty contracts above are the core runtime spine.

`contracts/` also carries harness and transport contracts that support the same spine without creating a second architecture:

- `Branch`
- `AgentJob`
- `AgentDispatchProfile`
- `AgentDispatchResult`
- `AgentJobReportResult`
- `AgentRunPacket`
- `ObservationRecord`
- `OpenThread`
- `ProviderSummary`
- `TelegramPacket`
- `AuthorityBoundary`
- `CapitalizationEntry`

These belong to the same contract axis.
They are harness surfaces around the runtime, not new kingdoms.

## Topology

The runtime converges on three runtime responsibilities:

1. `gateway`
2. `control-plane`
3. `worker`

This repository starts as a single binary and one codebase.
The topology stays conceptually split even if the process is not.

## Constitution Layout

- `constitution/`
  - active constitution for the new runtime
- `constitution/lineage.md`
  - records the distilled source lineage without creating a live dependency on an external repo
- current workspace path is `/Users/97layer/layer OS/constitution`; current runtime no longer depends on an external legacy `directives/` path
- shell-based absolute-path scanners must quote `/Users/97layer/layer OS/...` because the workspace root contains a space

## Package Map

- `cmd/layer-osctl`
  - CLI-first control surface for auth, preflight, policy, gateway, execute, verification, proposals, agent jobs, state, handoff, founder actions, work, flow, approval, release, deploy, rollback, and structure audit
- `cmd/layer-osd`
  - executable entrypoint that also serves the first founder cockpit
- `internal/runtime`
  - canonical types, validation, state, work, flow, approval, policy, execute, event, audit, system wiring
- `internal/api`
  - transport boundary, founder cockpit, snapshot read surface, audit read surface, adapter-capability surface, event SSE bus, and session bootstrap/finish surface
- `skills/active`
  - default skill grammar for the new runtime
- `skills/imported`
  - vendored skill library from the current Codex home

## Folder Checkpoint

```text
layer OS/
  constitution/
  contracts/
  docs/
  skills/
  cmd/
    layer-osctl/
    layer-osd/
  internal/
    api/
    runtime/
```

## Responsibility Axes

### Root Source Axes

1. `constitution`
2. `contracts`
3. `docs`
4. `skills`
5. `cmd`
6. `internal`

### Internal Runtime Axes

1. `api`
2. `runtime`

Runtime state is written to `.layer-os/` and is not treated as a source axis.
Write-auth token hashes are security control metadata rather than continuity state, so the default auth trust root lives in an external per-runtime auth file instead of `.layer-os/auth.json`.
The review room is runtime state in `.layer-os/review_room.json`, not a source markdown buffer.
A daemon-issued seal in `.layer-os/review_room.seal.json` guards that state. Direct rewrites of `review_room.json` are treated as integrity drift: audits surface them and the runtime opens a managed review-room integrity item instead of trusting the overwrite.
Event evidence is retained as a recent window plus `.layer-os/events_archive.json` for full history.
Adapter summaries expose gateway capability truth so agents do not infer unsupported provider fan-out.
Provider summaries also carry credential-ready truth internally, including
accepted env aliases such as `GEMINI_API_KEY` / `GOOGLE_API_KEY`, so daemon
dispatch and CLI council routing do not drift on provider readiness. Read-only
HTTP surfaces should redact auth-source/env-key metadata when projecting those
summaries outward.
Actor identity is treated as runtime data, not a kernel role binding: events prefer explicit `LAYER_OS_ACTOR`/request headers and otherwise fall back to neutral `system`, which keeps agent-roster changes detachable.
ProposalItem is the planning-stage registry before WorkItem: ideas enter runtime as proposals, stay visible in founder waiting lanes, and can later be promoted into discover-stage work without losing their evidence chain.
AgentJob is the first orchestration registry: queue planner/implementer/verifier work as runtime contracts, keep them founder-visible as waiting/now/risk lanes, persist them as `.layer-os/agent_jobs.json`, and let failed jobs auto-promote into review-room instead of spawning side-channel task files.
Terminal agent reports now close through `/api/layer-os/jobs/report`, emitting `agent_job.<status>` events and appending capitalization entries so the corpus, review-room, and founder lanes stay on the same Go runtime path. Optional `ChainRules` in `job.payload` can now create the next agent lane after a terminal report, keeping multi-agent follow-up inside canonical runtime contracts instead of side-channel scratchpads.
ObservationRecord is the raw intake registry for cross-channel context before promotion: `/api/layer-os/observations` / `layer-osctl observation` persist normalized captures with source channel, topic, refs, confidence, raw excerpt, and summary, then expose them for later linking/deduping without letting every note mutate founder memory directly.
The default single-thread intake lane now sits one step earlier: `/api/layer-os/conversation` / `layer-osctl conversation note` accept a single message with source, refs, tags, and toggle overrides, mask sensitive text, persist the note as an observation, and only escalate into proposal/job/review-room when the rule lane says it should.
Terminal-first auto-ingest seeds that registry from `session.finish` and `job.report`, and the operator-facing capture lanes (`layer-osctl ingest telegram`, `layer-osctl ingest content`, founder Telegram `/note`, founder Telegram DM free-text) now enter the same conversation spine instead of bypassing straight into raw observations.
Agent dispatch also separates transport truth from baton truth: a real provider or agent HTTP send stays `gateway.status=sent`, while missing provider endpoints downgrade to a recorded `job_packet` baton (`dispatch_state=packet_ready`) so official handoff can continue without pretending provider fan-out happened.
For bounded single-thread multi-model work, `job.payload.council` can request a sequential provider pass inside one runtime dispatch. The runtime still records one `GatewayCall` per provider, then projects the selected primary provider back into the top-level job result so the baton stays compatible with normal lanes. Native `quickwork` may auto-attach that council for bootstrap-derived planner/designer lanes, and daemon dispatch now also auto-upgrades non-manual planner/designer lanes into the same bounded council when more than one direct provider is actually dispatch-ready. The same quickwork loop can also promote one bounded context job out of `open_threads` / `parallel_candidates` before falling back to manual founder input or idle exit. When a proposal candidate already owns an `open_thread`, context promotion now leaves that thread alone so the founder sees one planning surface instead of a duplicated planner job. That keeps founder loops packet-first instead of payload-handwritten while leaving explicit founder-manual jobs single-provider by default.
The agent run packet now also carries a typed `runtime` contract (`source_of_truth`, `dispatch_transport`, `report_path`, `report_command`, terminal statuses, and write-auth env when required) so Python or other external wrappers can rejoin the official Go lane through `job.report` instead of inventing side-channel completion rules. `dispatch_transport=job_packet` remains the pull path, while daemon-pushed handoff uses `dispatch_transport=http_push`.
Next-agent handoff now also has a dedicated packet: `/api/layer-os/jobs/packet` / `layer-osctl job packet` bundle the assigned job with knowledge, handoff, and any linked proposal so baton passes stop depending on ad-hoc thread context. When a follow-up lane is auto-dispatched, the daemon can now push that same packet to a configured endpoint instead of stopping at `packet_ready`.
Review room is also the runtime meeting secretary: structured agenda items carry text, kind, severity, source, optional refs, optional `why`, optional `why_unresolved`, optional `contradiction`, optional `contradictions`, optional `pattern_refs`, and timestamps; high-signal failures are auto-promoted into open agenda items, then moved through `accept`, `defer`, or `resolve` as decisions harden. Open agenda insertion also checks accepted decisions for simple keyword overlap and records matched decision text into `contradictions` so contradiction review stays attached to the agenda itself. Defer transitions now also auto-stamp `why_unresolved` from the resolution reason so deferred tension does not collapse back into a bare todo. Snapshot export/import now also carries full review-room state so founder agenda continuity survives handoff and recovery.
That secretary loop now includes rejected approvals, carry-over risk at session finish, and auto-resolution of stale approval-rejection agenda when approval later succeeds.
Founder summary is downstream of that meeting surface, so open agenda items outrank passive waiting cards in handoff and cockpit summaries.
Handoff additionally surfaces structured secondary parallel candidates so agent workers can pick off non-primary work without inventing their own queues.
KnowledgePacket is the thinner sibling to handoff: it keeps only the current focus, short next-step/risk lists, current founder priority, top open agenda, the top 2-3 related corpus lessons, a small `surprising` list for mismatch/stall signals, a small `observation_links` list for repeated ref/semantic clusters, a small `proposal_candidates` list for gated pre-proposal promotion with explicit create/promote helpers driven by repeated evidence and high-signal review refs, a small `open_threads` list for structured unresolved questions, a small `action_hints` list for derived operator prompts, a small `action_routes` list for machine-usable routing, top parallel lanes, and capability hints for low-token context injection. That `surprising` lane is now stage-1 rule-based corpus synthesis: repeated focus, repeated agent failures, and persistent review carry-over surface as questions before they become new state. `open_threads` is the derived understanding surface: it does not own state, it projects unresolved questions out of corpus patterns and review-room tension.
Prompt work follows the same split: stable behavior belongs in `AGENTS.md`/constitution/docs, lane-specific prompt intent belongs in proposals and agent jobs, and baton-pass context belongs in official packets rather than ad-hoc thread memory. See `docs/prompting.md`.
Session bootstrap/finish sits on top of that thinner packet: bootstrap gives every entry surface the same read-only starting projection, while finish closes a session by updating memory and emitting a `session.finished` event into the shared bus.
That same finish path can also derive a secondary capitalization ledger in `.layer-os/knowledge/corpus/`, where situation/decision/cost/result facets become machine-readable business intelligence without giving the corpus its own write authority.
The event surface is both evidence and transport: append-only events stay in runtime history while `/api/layer-os/events/stream` lets cockpit and future agents react without inventing a second queue. Recent events stay in memory; archive evidence streams to disk and is reloaded only for explicit full-history paths such as snapshot/export.
Telegram remains bounded in this phase: the runtime can derive a compact `TelegramPacket` read model for preview and delivery, founder route commands can write limited runtime decisions such as source-intake routing, approval resolution, review-room transitions, and job dispatch, and founder DM free-text now persists into the canonical conversation spine. Clear founder DM operator intent should execute through those canonical runtime paths instead of roleplaying completion in chat. Telegram is still a control and intake surface, not a general-purpose editor or raw data lake.
WriteLease remains global at the file-backed persistence boundary: `persistLocked` saves a multi-file runtime snapshot as one batch, so cross-process domain lease splitting would weaken integrity before it improves concurrency. Parallelism belongs inside the single daemon queue and proposal/review/session lanes until storage stops depending on one atomic persistence sweep.
The canonical continuity host is a 24/7 VM runtime. Founder laptops are burst workstations for local development, verification, and recovery drills, not the always-on source of company continuity. Containerization stays deferred while binary-first deployment keeps the runtime simpler than the environment drift it would solve.
Daemon heartbeat is intentionally separated from company state: `/healthz` and `/api/layer-os/daemon` expose startup time, uptime seconds, and degraded reasons without leaking process-local uptime into snapshot or handoff packets.
Soul-Contract stays asymmetric: `Soul` is an interpretive operator layer for persona, tone, framing, and founder-facing meaning, while `Contract` remains the only executable state boundary. Soul may shape proposal summaries, memory, knowledge, handoff, and founder-facing language, but it does not get an independent persistence axis, write path, or execution authority. If Soul wants to change runtime, it must project into canonical contracts first. See `constitution/soul.md` for the continuity boundary that sits above packet posture.
Environment pressure is modeled as advisory, not a second control plane: founder-summary and knowledge expose host class, power mode, continuity role, recommended agent mode, and a founder-facing notice so low-power burst hosts can demote heavy work before they become silent runtime bottlenecks.
MCP tooling is also advisory, not a runtime dependency: Layer OS authority stays in Go contracts, CLI commands, daemon routes, and packet/report flows. Optional MCP servers may improve local ergonomics, but missing MCP must not block the core terminal-first harness. `layer-osctl audit mcp` and `./scripts/check_codex_mcp.sh` should be read as optional environment diagnostics, not startup gates.
Dynamic payload surfaces are kernel-validated recursively so snapshot imports and in-process callers cannot smuggle non-JSON runtime values into persisted state.

## GatewayCall Role

`GatewayCall` records the Go runtime's intent to dispatch to a provider or agent.
It is not a record of what happens inside an external agent.

- If a real provider endpoint is reachable, the call transitions to `status=sent`.
- If no provider endpoint is configured, the call stays `status=recorded` and
  `dispatch_state=packet_ready` marks that a baton was issued without a live send.
- LLM calls made by a Python agent internally are opaque to the Go runtime.
  Those calls are the agent's own responsibility and do not produce `GatewayCall` records.

The boundary: Go runtime → agent handoff is `GatewayCall`.
Agent → provider is outside the runtime's record surface.

See `docs/agent-integration.md` for the full agent-side contract.

## Design Rules

- Constitution is copied, not rewritten.
- Active constitution is distilled from source, not duplicated from it.
- Runtime is standard-library-first.
- Skills are CLI-first.
- Imported skills are vendor assets; active skills are the default entrypoints.
- Dynamic plugin behavior is avoided.
- Naming is part of architecture, not cosmetics.
- New folders require a clear responsibility that cannot fit an existing package.
- File-level modularity is preferred before folder-level expansion.
- The structure should optimize for fewer exceptions and faster interpretation, not for theoretical purity.
