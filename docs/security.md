# Security Posture

## Intent

Layer OS keeps security as an always-on operational lane, not an event-only patch lane.

## Baseline

- Mutating CLI/API paths must remain bearer-gated when write auth is enabled.
- Cross-origin writes are rejected; same-origin write protection is part of the runtime boundary.
- Repeated write-auth failures, cross-origin write probes, and non-loopback bootstrap attempts should emit `security.*` runtime events and promote review-room items when thresholds trip.
- `audit security` should summarize recent suspicious security signals so operators can tell whether the daemon has seen unusual agent or client behavior in the last day.
- Cockpit and API responses must keep protective headers active.
- HTTPS responses should emit HSTS and the API should keep restrictive browser security headers such as `Permissions-Policy`, `Cross-Origin-Opener-Policy`, and `Cross-Origin-Resource-Policy`.
- Provider egress must stay restricted to public `http`/`https` endpoints that pass runtime validation.
- Runtime state under `.layer-os/` must not become a plaintext secret sink.
- The write-auth trust root should live outside `.layer-os/` by default so runtime-data rewrites do not also own auth verification state.
- Initial write-auth bootstrap should be loopback-only until a bearer token is configured.
- External exposure requires edge TLS and edge access control outside the daemon.

## Operational Cadence

Run `go run ./cmd/layer-osctl audit security --strict`:

- weekly
- before every release
- before any external exposure

A passing security review preflight must record at least these checks:

- `write_auth_enabled`
- `secret_plaintext_surface_minimized`
- `edge_tls_required`
- `edge_access_control_required`

## Authority Boundary

- Review-room state lives in `.layer-os/review_room.json` but must be mutated only through `layer-osd` review-room transitions.
- External agent completion closes only through `layer-osctl job report ...` or `/api/layer-os/jobs/report`.
- Read-only bootstrap surfaces (`knowledge`, `handoff`, `session bootstrap --allow-local-fallback`) do not create write authority.

## Secret Handling

- Keep provider/API secrets in environment or external vault tooling; do not persist plaintext secrets into runtime JSON.
- Write-auth hash storage may be overridden with `LAYER_OS_AUTH_FILE`, but the secure default is an external per-runtime auth file rather than `.layer-os/auth.json`.
- If runtime secret plaintext findings appear in `audit security`, treat that as a release/exposure blocker.

## Exposure Gate

The daemon is not its own edge. If Layer OS is exposed beyond localhost, the operator must provide:

- TLS termination
- access control / identity gate
- explicit write token handling
- routine security review evidence
