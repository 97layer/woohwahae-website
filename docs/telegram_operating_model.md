# Telegram Operating Model

## Intent

Keep Telegram useful as an operating surface without repeating the old pattern
where one bot and one chat silently became founder assistant, ingestion inbox,
ops pager, approval room, and publish log all at once.

## Current Decision

Use **one internal bot with multiple chats**, not multiple bots.

Canonical routes:

- `founder`: founder room packets, decision nudges, and the canonical founder alert lane
- `ops`: deploy, job, daemon, automation, and failure notices
- `brand`: content review, approval, publish receipts, brand-source questions

The current internal operating bot is `@layer_os_staff_bot`.

Inbound policy is also route-scoped now:

- founder route: founder room by default, with free-text assistant reserved for founder 1:1 DM
- founder route also owns source-intake routing commands like `/intake`, `/route`, and `/redraft`
- ops route: lightweight read commands such as `/status` and `/review`
- brand route: lightweight read/review commands only until brand workflows land
- any chat: `/whoami` can reveal the current `chat_id` for route wiring
- unmapped chats: setup guidance plus `/whoami`, not full assistant behavior

## Why This Split

The legacy Telegram stack around the old public bot identity was heavily
blended. Reverse-tracking the old codebase showed that Telegram was acting as:

- an executive secretary / second-brain surface
- a freeform intake path for text, URLs, YouTube, images, and PDFs
- an approvals and publish trigger lane
- an ops and reminder lane
- a client workflow lane

Key evidence used to live in the legacy reference workspace across:

- the deleted Python Telegram secretary stack
- deleted conversation/council/content-publisher modules
- deleted knowledge ledgers for conversation context and long-term memory
- deleted raw Telegram signal caches
- deleted approval inbox and secretary logs
- deleted env snapshots outside git

Important truth from that scan:

- the exact handle `@official_97Layer_OSwoohwahae_bot` was not found in source control
- the old public bot identity was likely configured outside git
- legacy Telegram state is spread across conversations, signals, files, logs,
  approvals, and published assets

That means we should **not** try to reproduce the old Telegram behavior by
copying a single bot identity forward.

## Recommended Route Map

### Founder Route

Use for:

- direct founder DM
- assistant replies
- founder alert packet
- short decision nudges
- compact “what matters now” summaries

Do not use for:

- noisy job spam
- content proposal firehose
- raw scrape dumps

Env:

- `TELEGRAM_FOUNDER_CHAT_ID`
- `TELEGRAM_FOUNDER_DM_CHAT_ID`
- legacy fallback only: `TELEGRAM_CHAT_ID`

Inbound:

- founder room keeps alerts, decisions, and review pressure in one place
- founder 1:1 DM is the preferred surface for free-text assistant replies
- founder 1:1 DM free-text also persists into the canonical conversation spine, so important asks can open proposal / planner / risk lanes instead of dying as chat-only context
- founder 1:1 DM also acts on clear operator intent without forcing slash commands first: approval resolution, job dispatch, and review-room transitions should route through canonical runtime actions instead of roleplaying completion
- `/handoff` and `/note` are treated as founder DM behaviors, not founder room behaviors
- `/drop <링크나 텍스트>` stores a new raw source unit directly from Telegram
- founder room plain messages that clearly look like source material
  (link, long pasted text, multiline note) can also auto-ingest without the explicit `/drop`
- `/intake` lists recent unresolved source-intake items waiting for a route decision
- `/route <observation_id> <97layer|woosunhokr|woohwahae|hold>` records the founder route decision, and non-`hold` choices now auto-open an account-specific draft-seed proposal
- `/approvals`, `/approve`, `/reject`, `/jobs`, `/dispatch`, `/accept`, `/defer`, `/resolve` expose the bounded founder control set when explicit command syntax is better than free text
- `/drafts` lists the latest open draft seed for each source/account pair so founder can see the current working text without digging through observations
- `/redraft <draft_observation_id> <메모>` reopens a seed as a new revision and returns a fresh preview without mutating the earlier draft in place
- `/prep <draft_observation_id>` opens the canonical Threads prep lane from the current draft seed so approval/publish can stay on the same backend spine
- `/whoami` returns the current chat metadata so founder room and founder DM wiring are both self-serve
- if founder is currently pointed at a group room, normal free-text replies stay off until a dedicated DM chat id is seeded

### Ops Route

Use for:

- release / deploy / rollback notices
- job dispatch failures
- worker escalation
- daemon degradation
- runtime integrity or provider issues

Do not use for:

- everyday founder coaching
- brand review discussion

Env:

- `TELEGRAM_OPS_CHAT_ID`

Current behavior:

- if ops chat is missing, ops notices may temporarily fall back to founder
- this fallback is acceptable for bootstrap, not as the steady-state target
- free-text assistant replies are intentionally blocked here so ops traffic does
  not turn back into a founder side-channel
- `/whoami` makes it easy to capture the final group `chat_id` without spelunking logs

### Brand Route

Use for:

- content proposals
- approval prompts
- publish receipts
- brand-source review
- creative feedback loops

Do not use for:

- low-level infrastructure noise
- raw external scrape firehose

Env:

- `TELEGRAM_BRAND_CHAT_ID`

Brand is intentionally optional until content review actually starts living in
Telegram.

Inbound:

- free-text assistant replies stay blocked here for now
- keep the room focused on review, approval, and publish traffic
- `/whoami` is still allowed so brand room wiring is easy to finish

## What To Keep Out Of Telegram

Telegram should be the summary, approval, and control layer. It should not
become the raw data lake.

That control layer also should not be a hollow shell. Founder DM can talk
freely, but the meaningful parts of that exchange should land in canonical
runtime state so Telegram remains an actual operating surface instead of a
disconnected chat window.

It also should not fake completion. If Telegram did not actually resolve an
approval, dispatch a job, or move a review item through the canonical runtime,
the assistant must not speak in completed-action language.

Keep these out of Telegram by default:

- full scrape results
- bulk media archives
- noisy logs
- long-form notebook exports
- repeated cron chatter

Those belong in runtime state, storage, or the web admin surface. Telegram gets
the distilled version and the decision hook.

## Migration Notes

1. Keep `@layer_os_staff_bot` as the internal operating bot.
2. Treat the old public bot identity as reference-only until its actual external
   usage is confirmed.
3. Seed `TELEGRAM_FOUNDER_CHAT_ID` first so founder delivery becomes real.
4. If the founder room is a group, seed `TELEGRAM_FOUNDER_DM_CHAT_ID` next so personal assistant replies move back to the 1:1 DM.
5. Add `TELEGRAM_OPS_CHAT_ID` next so ops notices stop falling back to founder.
6. Add `TELEGRAM_BRAND_CHAT_ID` only when content review and publish flow need a
   dedicated room.
7. Move any future ingestion or social polling output into jobs, summaries, and
   approval packets rather than direct chat spam.
8. Keep source-intake routing small: Telegram should decide the destination seat,
   not become the raw data lake itself.

## Live Runtime Contract

The current runtime now exposes route-aware Telegram status:

- founder delivery
- route list with `founder`, `ops`, `brand`
- per-route delivery state
- short notes about fallback or legacy alias use

This keeps the admin web, VM inventory, and daemon behavior aligned around one
model.
