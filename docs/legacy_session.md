# Legacy Session Port

## Intent

This note ports only session semantics from the legacy workspace.
Runtime authority stays with `layer-osd`, `layer-osctl`, and `/api/layer-os/*`.
It does not restore legacy file-preload or shell-handoff mechanics.

See `docs/legacy_inventory.md` for the baseline keep/port/drop inventory.

## Keep

- Start each session from a compact bootstrap packet, not from scattered operator memory.
- Finish each session by writing current focus, next steps, open risks, and evidence together.
- Keep a thin default bootstrap and a richer opt-in expansion path.

## Port Carefully

- Map the old preload checklist into `SessionBootstrapPacket` fields.
  - always present: `knowledge`
  - opt-in on `--full`: `handoff`, `review_room`, `capabilities`
- Check `layer-osctl daemon status`, `GET /api/layer-os/daemon`, or `GET /healthz` before using daemon-unavailable fallback. If the daemon remains unreachable, `layer-osctl session bootstrap --allow-local-fallback` is read-only recovery only and must stay `read_only=true` and `degraded=true`.
- Keep single-writer caution as `WriteLease` visibility and review-room integrity audit, not as a separate lock folklore.

## Drop

- Direct reads of `.layer-os` files as the live operator protocol.
- Shell handoff scripts or asset-registration scripts as canonical session close.
- Any second authority path that bypasses daemon/API writes.
- Any role-pipeline carryover from the legacy workspace.

## Legacy Mapping

| Legacy semantic | Current home | Decision |
|---|---|---|
| session start loads compact current state | `layer-osctl session bootstrap`, `GET /api/layer-os/session/bootstrap` | keep |
| session close leaves baton state and evidence | `layer-osctl session finish`, `POST /api/layer-os/session/finish`, `session.finished` event | keep |
| optional deep preload | `--full` packet expansion | port carefully |
| daemon unavailable read fallback | `layer-osctl daemon status`, `GET /api/layer-os/daemon`, `GET /healthz`, then `--allow-local-fallback` | port carefully |
| file-by-file preload checklist | none | drop |
| shell append handoff | none | drop |

## Current Packet Shape

`SessionBootstrapPacket` carries `source`, `read_only`, `degraded`, and `knowledge`, with optional `handoff`, `review_room`, and `capabilities`.

`KnowledgePacket` already gives the compact session lane:

- current focus and goal
- next steps and open risks
- review-room priority signals
- corpus lessons and observation links
- proposal candidates and open threads
- actor/provider registry hints

`session finish` atomically updates memory, appends `session.finished`, and can open review-room follow-up when unresolved risk must survive the handoff.

## Operator Rules

1. Start with `layer-osctl session bootstrap`.
2. Add `--full` only when the thin packet is insufficient.
3. If `layer-osd` appears unreachable, check `layer-osctl daemon status`, `GET /api/layer-os/daemon`, or `GET /healthz` first. Use `--allow-local-fallback` only when the daemon is still unreachable and only for read context.
4. Close with `layer-osctl session finish --focus ... --steps ... --risks ...`.
5. Do not treat `.layer-os/review_room.json` or any other runtime file as a user-facing session API.
