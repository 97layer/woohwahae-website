# Linux VM Web Bootstrap

## Intent

Deploy the founder/admin web surface onto the same Ubuntu continuity host as
`layer-osd` without turning the VM into a permanent build box.

## Related Docs

- `docs/brand-home/README.md`
- `docs/linux_vm_bootstrap.md`
- `docs/operator.md`

## Related Artifacts

- `scripts/deploy_brand_home_vm.sh`
- `scripts/install_public_edge_vm.sh`
- `scripts/issue_public_tls_vm.sh`
- `scripts/switch_woohwahae_dns.sh`
- `scripts/cloudflare_certbot_hook.sh`
- `scripts/vm_runtime_inventory.sh`
- `scripts/seed_vm_providers.sh`
- `scripts/sync_vm_ai_cli_auth.sh`
- `scripts/discover_telegram_chat_id.sh`
- `scripts/discover_vm_telegram_chats.sh`
- `scripts/install_vm_tmux.sh`
- `scripts/prune_vm_host.sh`
- `scripts/install_vm_ai_clis.sh`
- `scripts/vm_tmux_attach.sh`
- `scripts/vm_trim_legacy_runtime.sh`
- `scripts/systemd/layer-os-web.service`
- `scripts/systemd/layer-os-tmux@.service`
- `scripts/systemd/layer-os-web.env.example`
- `scripts/nginx/layer-os-web.local.conf.example`
- `scripts/nginx/nginx.layer-os.conf.example`
- `scripts/nginx/woohwahae-public.conf.example`
- `scripts/nginx/woohwahae-public.tls.conf.example`

## Shape

- runtime host: `97layer-vm`
- web release root: `/srv/layer-os/web/releases/<stamp>`
- web current symlink: `/srv/layer-os/web/current`
- bundled Node runtime: `/srv/layer-os/node/current`
- service: `layer-os-web.service`
- local bind: `127.0.0.1:3081`

## Canonical Bootstrap

From the local repo:

```bash
./scripts/deploy_brand_home_vm.sh --host 97layer-vm
```

The script:

1. builds `docs/brand-home` in Next standalone mode
2. reuses the matching Linux Node runtime already on the VM, or downloads it only on first boot
3. syncs `.next/standalone`, `.next/static`, and `public`
4. syncs the service/env/nginx deploy assets needed on the VM
5. installs `layer-os-web.service` on the VM
6. seeds `/etc/layer-os/layer-os-web.env` on first boot, but keeps the existing file on later deploys
7. restarts the remote web service
8. retries `/admin/login` and `/api/public/proof` over localhost until Next is ready

Use `--check` when you only want to validate the local build and release bundle:

```bash
./scripts/deploy_brand_home_vm.sh --host 97layer-vm --check
```

To see what is actually left running on the VM after deploy:

```bash
./scripts/vm_runtime_inventory.sh --host 97layer-vm
```

That inventory now also shows provider readiness without printing secrets, so
you can tell at a glance whether Telegram is truly usable on the host:

- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `GOOGLE_API_KEY`
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_FOUNDER_CHAT_ID`
- `TELEGRAM_OPS_CHAT_ID`
- `TELEGRAM_BRAND_CHAT_ID`
- `THREADS_ACCESS_TOKEN`
- legacy `TELEGRAM_CHAT_ID`
- derived `inbound_mode` (`off`, `command_only`, `assistant`)
- derived `founder_delivery`, `ops_delivery`, and `brand_delivery`
- derived `threads_publish` (`disabled`, `ready`)
- whether the founder route is still relying on the legacy alias

It also prints the effective `provider_env_path`, which matters if the daemon is
pointed at a non-default provider file.

To check or seed provider secrets onto the VM without hand-editing the file over
SSH:

```bash
./scripts/seed_vm_providers.sh --host 97layer-vm
./scripts/seed_vm_providers.sh --host 97layer-vm --apply
```

The script reads local shell env or the macOS Keychain entries used by
`scripts/start.sh`, updates `/etc/layer-os/providers.env` on the VM, restarts
`layer-osd`, and prints the most relevant post-apply runtime payload without
revealing the secrets themselves. For example, `THREADS_ACCESS_TOKEN` now
returns `/api/layer-os/social/threads` readiness instead of the unrelated
Telegram surface.

For the Threads lane specifically, the shortest secure path is:

```bash
security add-generic-password -U -a layer-os -s THREADS_ACCESS_TOKEN -w '<threads-token>'
./scripts/seed_vm_providers.sh --host 136.109.201.201 --user skyto5339_gmail_com --ssh-key "$HOME/.ssh/google_compute_engine_mobile" --keys THREADS_ACCESS_TOKEN --apply
```

To mirror those provider keys into the interactive user shell for `codex`,
`claude`, and `gemini` without hand-editing dotfiles:

```bash
./scripts/sync_vm_ai_cli_auth.sh --host 136.109.201.201 --user skyto5339_gmail_com --ssh-key "$HOME/.ssh/google_compute_engine_mobile" --apply
```

That bridge writes `~/.config/layer-os/ai-cli-auth.sh` on the VM, maps
`GOOGLE_API_KEY -> GEMINI_API_KEY`, and sources the file from the common shell
startup files so fresh tmux/SSH shells see the AI provider env automatically.

## Canonical tmux Cockpit

To keep one durable operator seat on the VM, install the canonical tmux cockpit:

```bash
./scripts/install_vm_tmux.sh --host 97layer-vm
```

This installs `tmux`, seeds `/etc/layer-os/tmux.env`, installs the boot-time
service `layer-os-tmux@<user>.service`, and creates a persistent session named
`layeros` for the SSH user. The default windows are:

- `ops` — shell in `/srv/layer-os/current`
- `status` — quick command hints and another shell
- `daemon-log` — `journalctl -fu layer-osd`
- `web-log` — `journalctl -fu layer-os-web`

Attach later with:

```bash
./scripts/vm_tmux_attach.sh --host 97layer-vm
```

Or directly:

```bash
ssh -t 97layer-vm 'tmux attach -t layeros'
```

This is the anchor for a later mobile `Termius + Tailscale` workflow, but it
already helps today by making the VM the stable operator seat even when the
local shell or chat thread changes.

Current known live state after installation:

- service: `layer-os-tmux@skyto5339_gmail_com.service` -> `enabled`, `active`
- session: `layeros`
- windows: `ops`, `status`, `daemon-log`, `web-log`, `dev`

## AI CLI Seat on the VM

To install the interactive coding CLIs on the VM without adding a second system
Node runtime, reuse the bundled Layer OS Node release:

```bash
./scripts/install_vm_ai_clis.sh --host 136.109.201.201 --user skyto5339_gmail_com --ssh-key "$HOME/.ssh/google_compute_engine_mobile"
```

This install path:

- reuses `/srv/layer-os/node/current`
- exposes `node`, `npm`, `npx`, and `corepack` through `~/.local/bin`
- installs `codex`, `claude`, and `gemini` into the user-local npm prefix
- updates the common shell startup files to source `~/.config/layer-os/ai-cli-path.sh`

Current known live state after installation on `2026-03-11`:

- `codex` -> `0.114.0`
- `claude` -> `2.1.72`
- `gemini` -> `0.33.0`
- `~/.config/layer-os/ai-cli-auth.sh` exists for `skyto5339_gmail_com`
- fresh login shells see `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, and `GEMINI_API_KEY`

Binary availability is solved by `install_vm_ai_clis.sh`; auth stays explicit
through the separate `sync_vm_ai_cli_auth.sh` bridge so the CLI shell mirrors
only the AI-provider subset instead of the entire provider env.

## Dev Workspace on the VM

If mobile or remote sessions need more than logs and light ops, bootstrap a
real development checkout that stays separate from the live release symlink:

```bash
./scripts/install_vm_dev_workspace.sh --host 97layer-vm
```

This creates `/srv/layer-os-dev`, syncs the current repo there, preserves the
live runtime path at `/srv/layer-os/current`, and refreshes the tmux cockpit so
it also exposes a `dev` window.

That means the practical split becomes:

- `/srv/layer-os/current` — live release, owned by deploy flow
- `/srv/layer-os-dev` — remote development workspace for hotfixes and mobile work

For mobile SSH clients like Termius, use the real host/IP, not the local macOS
alias `97layer-vm`. The current SSH target behind that alias is:

- host: `136.109.201.201`
- user: `skyto5339_gmail_com`
- port: `22`

The VM uses Google OS Login rather than a user-managed `authorized_keys` file.
For mobile clients, the current dedicated key path is:

- private key: `~/.ssh/google_compute_engine_mobile`
- public key: `~/.ssh/google_compute_engine_mobile.pub`

That key was added to OS Login on `2026-03-11` and verified directly against
`136.109.201.201` before the older exposed `google_compute_engine` key was
removed. If mobile SSH breaks again, rotate the mobile key through OS Login
instead of editing files on the VM.

Current known live state after the dev bootstrap:

- `/srv/layer-os-dev` exists
- `cmd/layer-osctl/main.go` is present in that checkout
- `./scripts/vm_tmux_attach.sh --host 97layer-vm --list` shows the `dev` window

The canonical route split is now:

- `TELEGRAM_FOUNDER_CHAT_ID` for founder room packets and canonical founder delivery
- `TELEGRAM_FOUNDER_DM_CHAT_ID` for founder 1:1 assistant replies
- `TELEGRAM_OPS_CHAT_ID` for deploy/job/runtime notices
- `TELEGRAM_BRAND_CHAT_ID` for content review and publish traffic

If your local machine still only has the old `TELEGRAM_CHAT_ID`, the seed
script will reuse it as the founder route during migration.

If the only missing field is the founder chat id, discover it from recent bot
updates first:

```bash
./scripts/discover_telegram_chat_id.sh
```

Send a message to the bot from the target founder chat, rerun the script, then
store that chat id locally as `TELEGRAM_FOUNDER_CHAT_ID` and apply it with
`seed_vm_providers.sh`.

Once the VM daemon owns polling, local `getUpdates` can return `409 conflict`.
In that case, use the daemon-backed journal helper instead:

```bash
./scripts/discover_vm_telegram_chats.sh --host 97layer-vm
```

That reads recent `layer-osd` journal lines and prints candidate chat ids seen
by the live polling loop, including inferred `route`, `chat_type`, and a
lightweight `chat_label` so you can tell DM vs group traffic apart without
logging the message text.

If the VM still has leftover legacy edge/runtime residue from the old stack:

```bash
./scripts/vm_trim_legacy_runtime.sh --host 97layer-vm --check
./scripts/vm_trim_legacy_runtime.sh --host 97layer-vm
```

For a broader host cleanup that keeps the active Layer OS stack but trims stale
releases and disposable caches:

```bash
./scripts/prune_vm_host.sh --host 136.109.201.201 --user skyto5339_gmail_com --ssh-key "$HOME/.ssh/google_compute_engine_mobile" --check
./scripts/prune_vm_host.sh --host 136.109.201.201 --user skyto5339_gmail_com --ssh-key "$HOME/.ssh/google_compute_engine_mobile"
```

That path preserves:

- `layer-osd`, `layer-os-web`, `nginx`, and the tmux cockpit
- `/srv/layer-os/current`
- `/srv/layer-os/web/current`
- `/srv/layer-os-dev`
- the bundled Node runtime and installed AI CLIs

It trims:

- stale legacy unit files from `/etc/systemd/system` after backing them up in
  the VM user home
- old app/web release directories outside the keep window
- apt cache, npm cache, pip cache, node-gyp cache, and oversized journal data

Current known live state after the cleanup on `2026-03-11`:

- root disk usage moved from `78%` down to `73%`
- `/srv/layer-os` shrank to about `550M`
- `/srv/layer-os-dev` remained intact at about `920M`
- legacy unit residue now reports `none`

To make the VM ready to receive the live `woohwahae.kr` brand domain from
Cloudflare:

```bash
./scripts/install_public_edge_vm.sh --host 97layer-vm
```

This installs a canonical nginx layout on the VM, restores `conf.d` loading,
and proxies `woohwahae.kr` / `www.woohwahae.kr` to `127.0.0.1:3081`.

To bind real HTTPS on the VM:

```bash
./scripts/issue_public_tls_vm.sh --host 97layer-vm
```

The TLS script reuses an existing `/etc/letsencrypt/live/woohwahae.kr/*`
certificate when present. If the certificate is missing, it falls back to a
Cloudflare DNS-01 flow using the local `/tmp/CLOUDFLARE_API_TOKEN`. The current
default domain set is `woohwahae.kr`, `www.woohwahae.kr`, and
`admin.woohwahae.kr`.

To switch the live Cloudflare DNS from the old Pages route to this VM:

```bash
./scripts/switch_woohwahae_dns.sh --delete-edgecheck
```

That DNS switch now keeps three surfaces aligned:

- `woohwahae.kr` -> proxied public apex on the VM
- `www.woohwahae.kr` -> proxied CNAME to the apex
- `admin.woohwahae.kr` -> proxied admin-only host on the same VM

## Initial Secrets

On first boot, the deploy script seeds `/etc/layer-os/layer-os-web.env` if it is
missing.

- `SESSION_HMAC_SECRET` is generated locally unless already exported
- `LAYER_OS_WRITE_TOKEN` is taken from the current shell or the macOS Keychain
- `LAYER_OS_ADMIN_PASSWORD_SHA256` defaults to the SHA256 of that same write
  token unless you export a different hash before deploy

That bootstrap choice keeps the first protected login and the first write path on
one shared secret. Rotate it later once the dedicated admin login secret is set.

## Reverse Proxy

Keep the web service on `127.0.0.1:3081` by default. The VM edge now terminates
real TLS in nginx and proxies back to the local founder/admin web process.
That keeps `layer-os-web` private even after the public brand cutover. The apex
host is reserved for public pages, while `/admin*` and `/api/admin*` are
redirected to `admin.woohwahae.kr`.

## Content Boundary

This deploy path ships the standalone server, the `public/` web assets, and the
normalized brand/content pack that the site actually renders. It does not ship
raw Figma/Notion/Drive source exports or other authoring artifacts.

The admin Telegram surface now mirrors the daemon's readiness model. That means
operators can see preview content, send readiness, inbound mode, and founder
delivery blockers from the web without SSH'ing into the box first.

The founder/admin web now also includes a **Brand publish** surface. It seeds a
text-first draft from the normalized brand pack, then opens the canonical
proposal/work/approval/flow corridor for founder review without inventing a
separate publishing subsystem on the VM.

That same flow can now advance one approved text channel all the way out: when
`THREADS_ACCESS_TOKEN` is present on the VM, the admin card can publish approved
Threads drafts through the daemon and store a canonical publish receipt back
into observations. The raw absorbed brand source stays in the repo-local
content pack; the VM only sees the normalized pack and the resulting publish
evidence.

Brand identity and social style are now intentionally split. The VM ships the
current normalized brand spine plus imported legacy-caption examples for
Threads/Instagram style, but it does not depend on live Instagram crawling. A
future Instagram import can replace the style profile layer without rewriting
the shell or publish routes.

Latest known live rollout after the Threads slice:

- daemon current: `/srv/layer-os/releases/20260310_165020`
- web current: `/srv/layer-os/web/releases/20260310_165309`
- admin login localhost probe: `200`
- `/api/layer-os/social/threads`: live and returning `publish_configured=false`
  until `THREADS_ACCESS_TOKEN` is seeded

Latest known provider state after the token attach on `2026-03-11`:

- local `THREADS_ACCESS_TOKEN`: `present`
- VM `THREADS_ACCESS_TOKEN`: `present`
- live `/api/layer-os/social/threads`: `adapter=threads_api`, `publish_configured=true`

First bounded live publish evidence after that token attach:

- publish path: canonical daemon route (`proposal -> work -> approval -> flow -> observation -> /api/layer-os/social/threads`)
- active token owner: `97layer`
- proposal: `proposal_brand_threads_20260311101757_97layer_raw_001`
- approval: `approval_brand_threads_20260311101757_97layer_raw_001`
- flow: `flow_brand_threads_20260311101757_97layer_raw_001`
- thread id: `18096817897781944`
- publish receipt observation: `observation_1773224287559130589`
- published at: `2026-03-11T10:18:07.559032997Z`
- native topic tag: intentionally left blank; topic selection stays manual for now

Latest known rollout after the account-routing slice on `2026-03-11`:

- daemon current: `/srv/layer-os/releases/20260311_104254`
- web current: `/srv/layer-os/web/releases/20260311_104402`
- brand/admin draft prep now carries explicit destination account metadata (`97layer`, `woosunhokr`, `woohwahae`) before publish
- both VM deploy scripts now accept `--user`, `--port`, and `--ssh-key`, so direct-IP deploys with the rotated mobile OS Login key no longer depend on the old local SSH alias

The operator-facing CLI seam for that same flow is now present too:

- `layer-osctl threads status`
- `layer-osctl threads publish --approval <approval-id>`

That means tmux/VM operators no longer need to go through the admin web just to
check whether the Threads flow is ready or to fire the first bounded publish
smoke once the token is present.

Latest known live rollout after the brand-bridge/style split slice:

- web current: `/srv/layer-os/web/releases/20260310_180939`
- localhost home probe: `200`
- `layer-os-web.service`: `active`
- curated legacy media now ships inside `docs/brand-home/public/assets/media/brand`
- generated style corpus now comes from `scripts/import_legacy_social_style.py`
  -> `docs/brand-home/content/social-style-examples.generated.js`

As of the latest live Telegram rollout, that surface reports the split route
truth directly and the founder/ops/brand routes are all ready.
The old local macOS launch agent `com.layer-os.daemon` has now been unloaded so
the VM owns Telegram polling again, and the bot can self-report chat identity
through `/whoami`.

The founder/admin review-room page also now splits fresh blockers from older
unresolved history. This does not mutate runtime review state; it simply keeps
the current work visible when historical architect-loop failures are still
present in the room.

The next cleanup slice now goes one step further: the founder/admin review-room
page separates active blockers, stale runtime residue, strategic backlog, and
other unresolved items. This keeps the old architect-loop pile visible without
letting it impersonate the current delivery lane.

The founder/admin overview now also includes a `Source intake` card.
This is the first safe intake seat for founder-dropped links and texts:

- it stores normalized `source_intake` observations through the canonical daemon route
- it keeps the default destination on `97layer`
- it deliberately comes before any broad crawler or route automation
- it is meant to feed a later Telegram route-approval loop instead of publishing directly
- once founder Telegram resolves the route, the same card now shows the opened
  `source_draft_seed` inline on the intake row, so the first account-specific
  paragraph is visible from admin without spelunking raw runtime state
- that same card can now also open `Threads prep` directly from the latest
  draft seed, which reuses the canonical publish corridor instead of inventing
  another side path
- if one intake is translated into multiple account seats, the card now keeps
  the latest `97layer / woosunhokr / woohwahae` draft seeds side by side under
  the same source row

Latest known live web rollout after the source-intake visibility slice:

- web current: `/srv/layer-os/web/releases/20260312_000643`
- `layer-os-web.service`: `active`
- localhost `/admin`: `307`
- localhost `/admin/login`: `200`
- review-room UI now labels `local fallback` explicitly when the page is reading cached local state instead of the live continuity host

Latest known live daemon rollout after the Telegram intake-to-draft visibility slice:

- daemon current: `/srv/layer-os/releases/20260311_145403`
- `layer-osd.service`: `active`
- health: `status=ok`
- founder Telegram route now supports `/drop <링크나 텍스트>`, `/intake`, `/drafts`, `/route <observation_id> <97layer|woosunhokr|woohwahae|hold>`, `/redraft <draft_observation_id> <메모>`, founder-room auto-intake for source-like plain messages, automatic account-specific draft-seed opening for non-`hold` route decisions, and short preview listing for those seed drafts
- non-`hold` route decisions now also create a real `source_draft_seed` observation with account-specific first-body text, and the founder reply includes a short preview instead of only a proposal id
- `/redraft` now reopens that seed as a new revision observation and keeps the latest draft visible in `/drafts`, so Telegram can stay a light decision/edit seat instead of becoming the whole editor
- founder/admin source intake can now promote that latest draft seed into `Threads prep`, which opens the canonical proposal/work/approval/flow lane with the draft/source refs attached
- when that promotion happens, the prep text is reshaped once more so the
  outward account voice stays intact while raw internal draft labels and
  seat-explainer paragraphs do not leak into the publish lane
- `deploy_brand_home_vm.sh` now resolves the actual local `standalone` and
  `static` artifact directories before rsync, so web deploys survive small
  `.next` layout shifts instead of failing on a hard-coded path assumption
- release-bundle cleanup now keeps deploy manifests under `.layer-os/` and excludes local residue (`knowledge/`, `임시.md`) from daemon deploys, which reduced live review-room drift to the remaining real `gateway.failed` and `agent_job.failed` blockers only
- those last four blockers were historical Gemini planner failures from before `GOOGLE_API_KEY` was seeded on the VM; after confirming current provider readiness, they were resolved through canonical review-room transitions, so live VM review-room is now clear (`open_count=0`)

The latest known live release pair is:

- daemon: `/srv/layer-os/releases/20260311_145403`
- web: `/srv/layer-os/web/releases/20260312_012326`

After the latest Telegram route rollout, the VM reports:

- `adapter=telegram_bot`
- `send_adapter=telegram_bot`
- `send_configured=true`
- `polling_configured=true`
- `inbound_mode=assistant`
- founder route=`ready`
- ops route=`ready`
- brand route=`ready`

The current live Telegram split on the VM is now:

- founder: `TELEGRAM_FOUNDER_CHAT_ID=<founder room chat id>`
- founder dm: `TELEGRAM_FOUNDER_DM_CHAT_ID=<founder 1:1 dm chat id>`
- ops: `TELEGRAM_OPS_CHAT_ID=-5060692298`
- brand: `TELEGRAM_BRAND_CHAT_ID=-5293295684`

`TELEGRAM_FOUNDER_CHAT_ID`와 `TELEGRAM_FOUNDER_DM_CHAT_ID`를 같은 값으로 두면 founder alert가 1:1 DM으로 새므로, split이 깨진다. 두 값은 반드시 달라야 한다.

## Architect Mode

The continuity VM should usually stay in **observe mode**, not self-dispatch
mode. That means:

- `LAYER_OS_ARCHITECT_AUTODISPATCH=false`
- `LAYER_OS_ARCHITECT_AUTOVERIFY=false`

This keeps the always-on host from generating repeated repair jobs and review
spam when no bounded execution worker is meant to be running there.

To inspect or apply that posture:

```bash
./scripts/tune_vm_architect_mode.sh --host 97layer-vm --check
./scripts/tune_vm_architect_mode.sh --host 97layer-vm --mode observe
```

The current live VM already has that observe posture applied. `/api/layer-os/daemon`
now reports:

- `architect.enabled=true`
- `architect.auto_dispatch=false`

So the continuity host stays observant and recoverable without continuing to
spray self-repair jobs into the review backlog.
