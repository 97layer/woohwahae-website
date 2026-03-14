# Parallel Milestone

Historical checkpoint archive only.
This document records a past rollout slice and should not be treated as active runtime authority or default read-path material.

## Intent

This document records the current parallel implementation milestone so work can
resume cleanly after terminal loss, agent restarts, or handoff.

## Related Docs

- `docs/architecture.md`
- `docs/linux_vm_bootstrap.md`
- `docs/operator.md`
- `docs/brand-home/README.md`
- `constitution/brand.md`
- `constitution/voice.md`
- `constitution/surface.md`

## Active Milestone

Visible first cut:

- founder admin shows runtime state and can open quickwork lanes
- founder admin can preview and send the Telegram alert packet
- founder admin can tell whether Telegram is truly ready before sending
- founder admin can read the VM / release / deploy / rollback corridor
- founder admin can open a first text-first social lane from the local brand pack into proposal / approval
- public home shows a branded shell with live proof from Layer OS
- backend worker loop stays the operational spine

This milestone does not try to ship VM cutover, social automation, or apps in
the same slice. It prepares the shell they will plug into later.

## Current Roadmap Checklist

Use this as the live control-tower checklist for the next few slices.

### Closed

- founder/admin web is live on the VM and split from the public apex
- Telegram founder / ops / brand routing is live on one bot
- first live Threads smoke is closed on `97layer`
- account routing is explicit: `97layer` raw, `woosunhokr` refined, `woohwahae` polished
- direct-IP deploy scripts now work with the rotated mobile SSH key
- VM tmux cockpit and dev seat are installed
- founder/admin now has a safe `source intake inbox` that stores normalized source units before any route or publish decision, and each source row now shows its opened draft seed inline when route resolution has already happened
- Telegram founder route now has the first source-intake loop: `/drop`, `/intake`, `/drafts`, `/route`, `/redraft`, founder-room auto-intake for source-like plain messages, and automatic account-specific draft-seed observation + proposal opening on non-`hold` route decisions
- founder/admin can now promote a chosen draft seed straight into a `Threads prep` lane, so source intake no longer stops at a preview shell
- founder/admin review-room now splits real blockers, stale runtime residue, and strategic backlog so old loop noise no longer masks the current lane

### Now

- normalize more founder-dropped links / texts into the inbox and let the tag vocabulary settle
- keep everything landing in `97layer` first unless clearly destination-bound
- turn Telegram into the actual route-approval surface for intake items
- let route decisions generate a real first draft seed body instead of only an empty shell
- keep review-room focused on live blockers; treat old architect-loop failures as residue, not as the main control surface
- turn the draft-seed bridge into account-specific publish quality instead of growing Telegram commands further

### Next

- open `source intake -> Telegram route approval -> account draft -> publish prep` loop
- make `97layer` the raw intake seat for broad domains:
  development, beauty, finance, humanities, systems, daily life
- let `woosunhokr` translate selected source into beauty-practice / craft notes
- let `woohwahae` receive only the reduced magazine-like public version
- open the first `AIOps` lane: review-room compression, daemon freshness checks, and route recommendation without removing founder control

### Later

- add official / authorized connectors before broad crawling
- add compliant public collector only after intake metadata and approval meaning are stable
- add blog / Shorts / long-form reuse on top of the intake and account-translation loop
- add topic recommendation only after enough account-specific post history exists
- turn `97layer` into the source-of-truth flywheel: outside material -> worldview extraction -> account translation -> long/short/blog reuse

### Explicitly Deferred

- no broad crawler that hits risky login / private / paywalled surfaces
- no “self IP first, figure it out later” scraping path
- no direct source -> publish shortcut that bypasses account translation
- no automatic topic tagging as the final authority; founder stays the last editor

Latest checkpoint: `2026-03-10`

- backend quickwork coherence landed
- public home BrandSourcePack landed
- founder admin Telegram preview/send surface landed
- local `layer-osd` reachable; current status is `degraded` on `daemon_source_freshness`
- live smoke readouts succeeded: Telegram preview works, quickwork status shows implementer/verifier running and planner stopped
- live release corridor is still empty: `target/release/deploy/rollback` all returned `[]`
- founder admin deploy visibility landed on top of existing cockpit + daemon read models
- current legacy VM (`layer97-nightguard`, internal `10.138.0.2`) still serves `api.woohwahae.kr` via nginx; public routes point to legacy ports `5000`, `5001`, `8082`, and `8083`
- safe first-stop candidates on the legacy VM are the background automation lanes (`97layer-code-agent`, `97layer-ecosystem`, `97layer-gardener`, `97layer-telegram`), not the public edge (`nginx`, `woohwahae-backend`, `cortex-admin`, gateway on `8082`)
- legacy VM drain executed on `2026-03-10`: backup stored at `/home/skyto5339_gmail_com/layeros-drain-backup-20260310_143504`, legacy services/timers disabled, public ports `80/443/5000/5001/8082/8000` drained, and Layer OS base paths created under `/srv/layer-os`, `/var/lib/layer-os`, `/var/log/layer-os`, `/etc/layer-os`
- Linux VM continuity-host assets landed: `scripts/deploy_vm.sh`, `scripts/systemd/layer-osd.service`, env examples, and `docs/linux_vm_bootstrap.md`
- first remote bootstrap succeeded: `97layer-vm` now serves `layer-osd` from `/srv/layer-os/current`, the active release symlink points at `/srv/layer-os/releases/20260310_061615`, and remote `/healthz` is reachable
- remote daemon posture was initially `degraded` only on `security_posture=degraded`; that gap is now closed after enabling remote write auth and recording `security_review_20260310_063125`
- remote `/api/layer-os/status` now returns `status=ok`, `security.status=ok`, and empty daemon degraded reasons
- local control daemon now registers the real `vm` deploy target to `/bin/bash /Users/97layer/layer OS/scripts/deploy_vm.sh --host 97layer-vm`
- first release/deploy evidence landed locally: `release_vm_rollout_20260310_061612` and `deploy_vm_rollout_20260310_061612` both exist, and the deploy finished `succeeded`
- the first deploy smoke exposed a real CLI issue: long-running deploy calls outlived the default 30s client timeout even though the daemon eventually recorded success
- `layer-osctl` now gives long-running deploy/rollback/founder-release routes a 5m timeout budget, with env overrides via `LAYER_OS_HTTP_TIMEOUT` and `LAYER_OS_LONG_HTTP_TIMEOUT`
- a second deploy smoke (`deploy_vm_timeout_verify_20260310_063332`) also finished `succeeded`, which confirmed the state path but exposed a second issue: the currently running local daemon still has a 30s HTTP write timeout and can drop the response with `EOF` before returning the finished deploy record
- code now raises `cmd/layer-osd` write timeout to 5m so long deploy responses can complete once the local daemon is restarted; until that restart, local status will continue to show `daemon_source_freshness=degraded`
- local `layer-osd` has now been restarted on `2026-03-10`, which cleared `daemon_source_freshness`; local security review `security_review_local_20260310_065003` is also recorded, so local `layer-osctl daemon status` now returns `status=ok`
- end-to-end response verification is closed: `deploy_vm_response_verify_20260310_065018` returned a normal succeeded JSON response through the local daemon after the timeout fixes on both client and server sides
- founder/admin deploy lane now surfaces continuity-host evidence directly: host, continuity status, latest deploy/release refs, deploy duration, remote release dir, and security review timing are all shown from the existing runtime read path
- founder/admin web VM deploy assets landed: `scripts/deploy_brand_home_vm.sh`, `scripts/systemd/layer-os-web.service`, `scripts/systemd/layer-os-web.env.example`, `scripts/nginx/layer-os-web.local.conf.example`, and `docs/linux_vm_web_bootstrap.md`
- founder/admin web is now deployed on `97layer-vm` at `/srv/layer-os/web/releases/20260310_072936`, with `layer-os-web.service` active and localhost probes returning `login=200` and `proof=200`
- the deploy script now carries its own service/env/nginx assets into the staged release and waits through Next startup instead of false-failing on an early health probe
- public proof no longer depends on the heavyweight `/api/layer-os/cockpit` route; it now reads lighter runtime surfaces (`status`, `founder-summary`, `review-room`, `verifications`, `auth`, `corpus/entries`, `observations`) with bounded 3s upstream timeouts so the public shell degrades instead of hanging
- VM runtime inventory tooling landed in `scripts/vm_runtime_inventory.sh`; the current remote snapshot shows the intended always-on listeners only on `127.0.0.1:17808` (`layer-osd`) and `127.0.0.1:3081` (`layer-os-web`)
- legacy runtime cleanup tooling landed in `scripts/vm_trim_legacy_runtime.sh`; it masks known legacy edge/runtime units and trims the retired `cloudflared tunnel --url http://localhost:5001` process
- after the cleanup run on `2026-03-10`, `nginx.service` is `masked/inactive`, failed services are `0 loaded units listed`, and the old `cloudflared` tunnel is gone
- the current VM-visible always-on Layer OS set is now explicit: `layer-osd` on `127.0.0.1:17808` and `layer-os-web` on `127.0.0.1:3081`, plus normal GCP guest/ops agents outside the app surface
- `woohwahae.kr` is confirmed as the live brand domain, while `woohwaha.kr` is not the active target
- public edge tooling landed in `scripts/install_public_edge_vm.sh`, `scripts/issue_public_tls_vm.sh`, `scripts/switch_woohwahae_dns.sh`, plus `scripts/nginx/nginx.layer-os.conf.example`, `scripts/nginx/woohwahae-public.conf.example`, and `scripts/nginx/woohwahae-public.tls.conf.example`
- the VM public edge is now live for the brand domain: `nginx` is active on ports `80/443`, the VM public IP is `136.109.201.201`, and direct origin probes return `200` for both `/` and `/admin/login`
- existing Let's Encrypt material for `woohwahae.kr` was reused on the VM; `certbot 5.3.1` on the host rejects `--manual-public-ip-logging-ok`, so the public TLS script now skips that flag and skips re-issuing when the cert is already present
- Cloudflare DNS is now switched off the old Pages CNAME: `woohwahae.kr` is a proxied `A -> 136.109.201.201`, `www.woohwahae.kr` is a proxied `CNAME -> woohwahae.kr`, and the temporary `edgecheck.woohwahae.kr` probe record has been removed
- external verification now passes: `https://woohwahae.kr/`, `https://www.woohwahae.kr/`, and `https://woohwahae.kr/admin/login` all return `200`
- the protected admin surface is now split onto `https://admin.woohwahae.kr`: Cloudflare has a proxied `A -> 136.109.201.201`, the shared origin certificate now covers `admin.woohwahae.kr`, apex `/admin*` and `/api/admin*` routes `308` to the admin subdomain, and `https://admin.woohwahae.kr/admin/login` returns `200`
- Telegram readiness now has an explicit runtime story: the daemon reports send readiness, inbound mode (`off`, `command_only`, `assistant`), founder delivery state, and short setup notes instead of a simple enabled flag
- founder admin now renders that same Telegram readiness story directly in the card, so operators can see whether polling, founder delivery, and Gemini-backed replies are actually available before hitting send
- VM inventory now includes provider readiness without printing secrets, including `OPENAI_API_KEY`, `GOOGLE_API_KEY`, `TELEGRAM_BOT_TOKEN`, `TELEGRAM_FOUNDER_CHAT_ID`, `TELEGRAM_OPS_CHAT_ID`, `TELEGRAM_BRAND_CHAT_ID`, legacy `TELEGRAM_CHAT_ID`, plus derived `inbound_mode` and route delivery states
- the Telegram card semantics were tightened after verifier review: `command_only` no longer masquerades as `noop`, and a dedicated admin-state test now locks the `polling only -> ready` progression in place
- the VM inventory now follows the daemon's effective provider env path and prints that path, so the operator script and the live daemon cannot silently read different secret files
- daemon deploy health checks now retry through cold-start timing, which removes the previous false-fail pattern where the service was healthy but the script exited early
- founder/admin web deploys now reuse the existing remote `layer-os-web.env` and the already-installed remote Node runtime, so repeat deploys no longer depend on re-entering the write token or redownloading Node
- live rollout for this slice succeeded on `2026-03-10`: daemon release `/srv/layer-os/releases/20260310_101719` and web release `/srv/layer-os/web/releases/20260310_102228` are both active on `97layer-vm`
- current live Telegram truth is explicit: the runtime and admin card now load correctly, the VM has `GOOGLE_API_KEY` and `TELEGRAM_BOT_TOKEN` seeded, so inbound mode is now `assistant`, but founder delivery is still blocked because the founder route chat id is missing
- provider-seeding helper landed in `scripts/seed_vm_providers.sh`; it can check local-vs-remote provider readiness or push locally available secrets to the VM without printing them
- chat-id discovery helper landed in `scripts/discover_telegram_chat_id.sh`; it can inspect recent bot updates and print candidate founder chat ids without exposing the bot token
- `discover_telegram_chat_id.sh` hit the expected `409 conflict` once the VM daemon took over polling, so a daemon-backed helper now exists in `scripts/discover_vm_telegram_chats.sh` and `layer-osd` logs inbound Telegram chat ids without logging the message text
- live provider readiness on `97layer-vm` now reads `GOOGLE_API_KEY=present`, `TELEGRAM_BOT_TOKEN=present`, `TELEGRAM_FOUNDER_CHAT_ID=missing`, `TELEGRAM_OPS_CHAT_ID=missing`, `TELEGRAM_BRAND_CHAT_ID=missing`, `inbound_mode=assistant`, `founder_delivery=chat_missing`
- the daemon-backed chat discovery helper currently returns `no_recent_chats_found`, which means the bot has not yet seen a founder-side inbound message on the live polling loop
- Telegram is now being normalized around one bot with multiple route chats instead of one overloaded chat id: founder, ops, and brand are explicit runtime routes, ops alerts can temporarily fall back to founder, and brand stays opt-in until content review moves into Telegram
- inbound Telegram is now route-aware too: founder keeps the free-text assistant and note lane, ops/brand stay scoped to lightweight read/review behavior, and unmapped chats get setup guidance instead of silently acting like founder
- the Telegram reply prompt has been tightened toward a practical aide voice so live chat replies stop sounding like raw system output or clipped translationese
- legacy Telegram evidence has been translated into a new operating-model doc so future migration work does not re-collapse founder, ops, approvals, and content traffic into one chat again
- live rollout for the Telegram route/tone slice succeeded on `2026-03-10`: daemon release `/srv/layer-os/releases/20260310_114331` and web release `/srv/layer-os/web/releases/20260310_114631` are active on `97layer-vm`
- external host checks still pass after that rollout: `https://admin.woohwahae.kr/admin/login` and `https://woohwahae.kr/` both return `200`
- live Telegram now reports the corrected adapter story on the VM: `adapter=noop`, `send_adapter=noop`, `polling_configured=true`, `inbound_mode=assistant`, founder route=`chat_missing`, ops route=`disabled`, brand route=`disabled`
- live chat discovery is still quiet after the rollout: `discover_vm_telegram_chats.sh --tail 120` returned `no_recent_chats_found`, so the bot has not yet seen a fresh founder/ops/brand message under the new logging shape
- the first text-first social lane is now wired into founder/admin as **Brand lane**: it seeds draft presets from `docs/brand-home/content/public-home-source.js` and opens `proposal -> work -> approval -> flow -> observation` over canonical runtime routes instead of a separate publishing subsystem
- the first real text-channel publish corridor is now implemented for Threads: approved brand-lane drafts can publish through the daemon route `/api/layer-os/social/threads`, and successful publishes write canonical `brand_publish_threads` observation receipts with container/thread ids
- Threads uses the same VM provider-env spine as Telegram instead of inventing a separate web-only secret path; `THREADS_ACCESS_TOKEN` is now part of `scripts/start.sh`, `scripts/seed_vm_providers.sh`, and `scripts/vm_runtime_inventory.sh`
- the current brand source boundary is deliberate: legacy source structure still feeds the normalized pack for speed, and cleanup/re-authoring is deferred until after the first publish loop is proven
- social generation is now being split into two layers on purpose: brand identity stays in the normalized pack, while channel tone now reads a generated legacy-caption corpus so future Instagram-native import can swap style without rewriting the product shell
- `scripts/import_legacy_social_style.py` now generates `docs/brand-home/content/social-style-examples.generated.js` from legacy memory, and founder/admin surfaces the example count plus sample excerpts for the Threads profile
- founder/admin review-room now separates `Current blockers` from older unresolved history, so open-count noise from the old architect loop no longer hides what is actively blocking the current lane
- Brand lane now carries imported legacy style example ids through the prep observation and Threads candidate/receipt view, so the team can trace which earlier caption references informed a draft before the first real publish loop
- VM operator continuity now has a concrete next seat: `scripts/install_vm_tmux.sh` + `scripts/vm_tmux_attach.sh` set up a boot-persistent `layeros` tmux cockpit on `97layer-vm`, which is the base for later Termius/Tailscale handoff without forcing a full control-tower migration yet
- that tmux cockpit is now actually installed on `97layer-vm`; `layer-os-tmux@skyto5339_gmail_com.service` is enabled and active, and the live `layeros` session currently exposes `ops`, `status`, `daemon-log`, and `web-log`
- next mobile-ready step is now implemented too: `scripts/install_vm_dev_workspace.sh` can bootstrap `/srv/layer-os-dev` as a real remote checkout, refresh the tmux cockpit to add a `dev` window, and keep the live release path separate from ad-hoc remote development
- that remote dev seat is now live on `97layer-vm`: `/srv/layer-os-dev` exists, `cmd/layer-osctl/main.go` is present there, and the live `layeros` tmux session now exposes `ops`, `status`, `daemon-log`, `web-log`, and `dev`
- mobile SSH hygiene is now reset: a dedicated OS Login key at `~/.ssh/google_compute_engine_mobile` was created on `2026-03-11`, direct SSH with that key to `136.109.201.201` succeeded, and the previously exposed `google_compute_engine` key was removed from OS Login
- the VM now also has a lightweight interactive AI CLI seat: `scripts/install_vm_ai_clis.sh` reuses `/srv/layer-os/node/current`, installs `codex`, `claude`, and `gemini` under `~/.local/bin`, and verified fresh login-shell PATH visibility on `97layer-vm`
- provider/auth residue is now partly normalized too: `scripts/seed_vm_providers.sh`, `scripts/vm_runtime_inventory.sh`, and `scripts/discover_vm_telegram_chats.sh` all understand explicit `--user/--port/--ssh-key` access, `ANTHROPIC_API_KEY` is part of the provider spine, and `scripts/sync_vm_ai_cli_auth.sh` now mirrors `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, and `GOOGLE_API_KEY -> GEMINI_API_KEY` into `~/.config/layer-os/ai-cli-auth.sh` for fresh VM shells
- `scripts/seed_vm_providers.sh` no longer assumes Telegram as the only post-apply check; it now probes the most relevant daemon surface for the selected secret, so `THREADS_ACCESS_TOKEN` seeding feeds back through `/api/layer-os/social/threads`
- live AI CLI auth is now ready on `97layer-vm`: local provider seeding pushed `OPENAI_API_KEY` and `ANTHROPIC_API_KEY` to the VM alongside the existing `GOOGLE_API_KEY`, and fresh login shells now see `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, and `GEMINI_API_KEY`
- host cleanup tooling is now also normalized: `scripts/prune_vm_host.sh` keeps the active Layer OS stack while trimming stale releases and disposable caches, and `scripts/vm_trim_legacy_runtime.sh` now removes backed-up legacy unit files instead of still treating `nginx` as disposable
- live cleanup on `2026-03-11` reduced the root disk from about `78%` to `73%`, shrank `/srv/layer-os` to about `550M`, kept `/srv/layer-os-dev` intact at about `920M`, vacuumed apt/journal/npm/pip/node-gyp caches, and cleared visible legacy unit residue from `systemctl list-unit-files`
- the founder/admin web was redeployed after that slice on `2026-03-10`; live `97layer-vm` web current is now `/srv/layer-os/web/releases/20260310_180939`, `layer-os-web.service` is active, and localhost `/` returns `200`
- live rollout for the Threads slice succeeded on `2026-03-10`: daemon release `/srv/layer-os/releases/20260310_165020` and web release `/srv/layer-os/web/releases/20260310_165309` are active on `97layer-vm`
- live daemon now serves `/api/layer-os/social/threads` and reports `adapter=noop`, `publish_configured=false` until `THREADS_ACCESS_TOKEN` is seeded on the VM
- as of `2026-03-11`, `THREADS_ACCESS_TOKEN` is now present both locally and on `97layer-vm`; the live Threads route reports `adapter=threads_api` and `publish_configured=true`, so the remaining step is the first bounded publish smoke
- the currently seeded live Threads user token resolves to username `97layer`, so the first smoke post should be treated as infrastructure validation rather than the brand-official launch moment
- live founder/admin web now includes the Threads publish surface and still returns `admin_login=200`; the current blocker is only provider readiness, not code rollout
- the first bounded live Threads smoke is now closed on `2026-03-11` through the canonical daemon route using the `97layer` token: proposal `proposal_brand_threads_20260311101757_97layer_raw_001`, approval `approval_brand_threads_20260311101757_97layer_raw_001`, and flow `flow_brand_threads_20260311101757_97layer_raw_001` produced thread `18096817897781944` with canonical publish receipt observation `observation_1773224287559130589`
- that first live smoke intentionally shipped with no native `topic_tag`; topic choice stays manual for now so account-native labels can settle before recommendation or automation is added
- repo support for optional Threads `topic_tag` is already implemented locally, but the current live VM publish succeeded on the older no-topic rollout, so the next web/daemon deploy is what will make that newer topic surface live
- draft routing now also carries the explicit destination account (`97layer`, `woosunhokr`, `woohwahae`) through the admin lane, prep observation, candidate list, and publish receipt so the shared worldview can still split cleanly into raw / refined / polished seats without accidental mixing
- the VM deploy scripts no longer depend on the stale local `97layer-vm` SSH alias alone; `scripts/deploy_vm.sh` and `scripts/deploy_brand_home_vm.sh` now accept `--user`, `--port`, and `--ssh-key`, so direct-IP deploys with `~/.ssh/google_compute_engine_mobile` work cleanly from the current operator seat
- that account-routing slice is now live on `97layer-vm`: daemon current is `/srv/layer-os/releases/20260311_104254` and web current is `/srv/layer-os/web/releases/20260311_104402`
- source intake architecture is now explicitly documented too: `docs/source_intake_operating_model.md` fixes the next build order as `manual drop / authorized connector / compliant public collector -> normalized source unit -> Telegram route approval -> account-specific draft`
- the first safe source-intake slice is now live in founder/admin: `/api/admin/runtime/source-intake` is backed by canonical observation routes, the admin overview has a `Source intake` card, and the latest live web release is `/srv/layer-os/web/releases/20260311_112604`
- the next bridge slice is now live too: founder Telegram route can create raw source units with `/drop`, list unresolved intake items with `/intake`, record route decisions with `/route <observation_id> <97layer|woosunhokr|woohwahae|hold>`, and auto-open account-specific draft-seed proposals for non-`hold` decisions
- founder/admin review-room triage is now live too: the page separates active blockers, stale runtime residue, strategic backlog, and other unresolved items, so old architect-loop noise no longer looks like the current delivery lane itself
- current founder/admin web release on the VM is `/srv/layer-os/web/releases/20260311_135230`
- current daemon release on the VM is `/srv/layer-os/releases/20260311_134651`
- live review-room on the VM is now clear (`open_count=0`): the last four Gemini planner/gateway failures were historical `no_google_api_key` residue and were resolved with explicit rationale once provider readiness was confirmed
- review-room UI now also warns when the page is reading local fallback state instead of live runtime state, so stale local backlog cannot masquerade as the continuity host truth
- founder Telegram route now writes a real `source_draft_seed` observation with account-specific first-body text when `/route` resolves to `97layer`, `woosunhokr`, or `woohwahae`; the reply now includes a short preview instead of only a proposal id shell
- founder Telegram route can now also list those opened seed drafts with `/drafts`
- founder Telegram route can now reopen those seed drafts with `/redraft <draft_observation_id> <메모>`, which creates a fresh revision observation instead of overwriting the original and keeps only the latest draft visible in `/drafts`
- founder/admin source intake can now open `Threads prep` directly from a draft seed and reuse the canonical `brand publish` proposal/work/approval/flow corridor with draft/source refs attached
- source draft seed generation now separates account voice earlier too: `97layer` keeps a raw maker-diary seat, `woosunhokr` tilts toward beauty-practice/craft, and `woohwahae` tilts toward a quieter public shell before publish prep even opens
- founder/admin source intake now shows multiple latest draft seeds under one source row when the same intake is translated into more than one account, so `원천 하나 -> 여러 번역 seat` is visible instead of collapsing to one preview
- current daemon release on the VM is now `/srv/layer-os/releases/20260311_235710`
- current founder/admin web release on the VM is now `/srv/layer-os/web/releases/20260312_000643`
- `go test ./...` passed
- `npm test` passed in `docs/brand-home`
- `npm run build` passed in `docs/brand-home`

## Lane Split

### Control Tower

Owns:

- milestone decisions
- hot seam approval
- integration order
- durable checkpoint notes

Avoids:

- deep frontend-only polish
- backend-only worker implementation details unless integration is blocked

### Backend Lane

Owns:

- `quickwork`
- `job work`
- worker orchestration
- admin runtime status surfaces
- tests for the quickwork/admin flow

Primary write set:

- `cmd/layer-osctl/**`
- `internal/runtime/**`
- `internal/api/**`
- `scripts/**`

### Frontend Lane

Owns:

- branded public home
- founder admin presentation polish
- normalized brand source pack for web consumption

Primary write set:

- `docs/brand-home/**`

## Guardrails

- one canonical state story
- one authenticated write path
- no direct `.layer-os/*.json` rewrites
- no parallel edits on hot seams without explicit integration review

Hot seams for this milestone:

- `internal/api/router.go`
- `contracts/*.schema.json`
- `.layer-os/**`
- continuity/session core files

## Current Implementation Direction

### Backend

- use the existing worker spine instead of inventing a new queue
- keep founder admin quickwork visibility coherent
- verify the CLI and web quickwork surfaces point at the same runtime story

### Frontend

- introduce a `BrandSourcePack`-style normalized content layer
- feed public home from that pack plus runtime proof
- keep tone quiet, exact, and founder-legible

### Brand Source

Source-of-truth is external in the long run.

For this milestone:

- normalize brand content into a web-consumable local pack
- keep the pack easy to replace from future external imports
- do not turn brand into a separate runtime authority

## Verification

Core commands:

```bash
go test ./...
```

```bash
cd docs/brand-home && npm test
```

Use runtime smoke when the daemon path is reachable and permitted:

```bash
go run ./cmd/layer-osctl quickwork --status
```

```bash
go run ./cmd/layer-osctl daemon status
```

```bash
go run ./cmd/layer-osctl threads status
```

```bash
./scripts/deploy_vm.sh --host 97layer-vm --check
```

```bash
./scripts/deploy_brand_home_vm.sh --host 97layer-vm
```

```bash
./scripts/vm_runtime_inventory.sh --host 97layer-vm
```

```bash
./scripts/vm_trim_legacy_runtime.sh --host 97layer-vm --check
```

```bash
./scripts/install_public_edge_vm.sh --host 97layer-vm
```

```bash
./scripts/issue_public_tls_vm.sh --host 97layer-vm
```

```bash
./scripts/switch_woohwahae_dns.sh --check
```

## Next Checkpoints

1. turn the new `Brand lane` prep corridor into one real publish corridor for a single text channel, with a publish receipt/log instead of manual-only prep
2. wire brand-source injection from external source into the normalized web pack
3. choose the first external brand-source importer and normalize it into the existing web pack shape
4. tighten the always-on production set on the VM to daemon + web + Telegram only, with all heavy workers left on-demand
5. move any remaining admin-only links and operator habits over to `admin.woohwahae.kr` so the apex stays purely public-facing
6. dedicated Telegram route wiring is now live end-to-end on the VM: founder DM=`7565534667`, ops=`-5060692298`, brand=`-5293295684`
7. `/api/layer-os/telegram` is back to full assistant mode with founder canonical delivery and founder DM both anchored to the 1:1 chat, while ops and brand stay on their own rooms
8. the always-on VM is now in architect observe mode (`LAYER_OS_ARCHITECT_AUTODISPATCH=false`, `LAYER_OS_ARCHITECT_AUTOVERIFY=false`) so it stops self-dispatching repair jobs before the first social lane opens
9. old review-room noise still needs a one-time hygiene pass, but the VM should no longer keep making the same architect-loop failure pattern worse
10. decide whether `OPENAI_API_KEY` is truly needed on the VM always-on path or should stay absent by design
11. `layer-osctl threads status|publish --approval <id>` now exists as the tmux-side seam for the first real social publish smoke, so the only remaining blocker for live publish is `THREADS_ACCESS_TOKEN`
12. turn the accumulating runtime trail into a founder-facing non-dev journal/blog source, so work logs, approvals, releases, Telegram summaries, and social receipts can later be rewritten into narrative updates instead of staying trapped as engineering residue
13. keep social accounts separated by source and tone: `97layer` for maker/process notes, `woosunhokr` for personal beauty craft, and `woohwahae` for the quieter official brand lane after more source material is collected
14. once Threads posting settles, attach account-native topic labels instead of one shared tag pool: start with `97layer` as `vibe coding / build log / system cleanup`, then define the refined and polished sets for `woosunhokr` and `woohwahae`
15. keep the deeper architecture explicit: `97layer / woosunhokr / woohwahae` are not three disconnected identities but one shared philosophical spine shown at different polish levels, with legacy anchors in `the_origin`, `sage_architect`, `about`, and the old `woosunho` page
16. long-term content engine should treat `97layer` as the raw intake hub for founder notes and external sources, then distill recurring worldview patterns and translate them outward into `woosunhokr`, `woohwahae`, blog, Shorts, and longer-form media as a one-source-multi-use flywheel
17. Threads publish now carries an optional native `topic_tag` from draft creation through approval, publish, and receipt so account-specific topic labels can join the live lane without inventing a side channel
18. connect the new source-intake inbox to Telegram route approval so founder can decide `97layer / woosunhokr / woohwahae / hold` from chat instead of only from the web card
19. keep tightening `source_draft_seed -> Threads prep` so the prep lane reflects the outward account voice rather than leaking raw internal draft labels and explainer sentences straight into publish
20. founder/admin web is now live again at `/srv/layer-os/web/releases/20260312_012326`, and the web deploy path was hardened to resolve actual local `.next/standalone` and `.next/static` artifact roots before rsync instead of assuming one fixed layout
