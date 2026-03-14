# Source Intake Operating Model

## Intent

Keep the source side simple enough to grow, while making sure three things do
not collapse into one mess:

- raw intake
- account translation
- publishing approval

This is the document for how outside material should enter the system before it
ever becomes a post.

## Core Idea

The mental model is not:

- crawler -> post

It is:

- source intake -> interpretation -> route proposal -> Telegram approval ->
  account-specific draft -> publish

That is the only way `97layer`, `woosunhokr`, and `woohwahae` can share one
spine without sounding merged.

## Intake Classes

### 1. Manual Drop

Use when the founder sends material directly.

Examples:

- Telegram link
- pasted text
- screenshot or file
- voice note
- note copied from another app

This is the safest and fastest intake lane.

Default destination:

- `97layer`

Why:

- founder intent is explicit
- no crawler ambiguity
- approval meaning is easy

### 2. Official / Authorized Connector

Use when the source offers a real API, export, or authorized integration.

Examples:

- Threads API
- YouTube API
- RSS feed
- site export
- owned-domain content feed

This is the preferred automation lane after manual drop.

Default destination:

- `97layer` first, unless the source is already clearly account-specific

### 3. Compliant Public Collector

Use only for public material that is allowed into the system's safe collector
lane.

Rules:

- no login wall
- no private content
- no paywall bypass
- no credentials borrowing
- no full-content mirroring by default
- keep provenance and fetch policy attached

This lane should produce:

- URL
- title
- short excerpt
- summary
- source policy

Not a giant unstructured dump.

## Red / Yellow / Green Source Policy

### Green

Okay to automate first:

- founder-provided material
- owned sites
- official APIs
- RSS / export / sitemap on allowed sources
- open/public material with clear reuse posture

### Yellow

Collect carefully, then require founder approval before any draft opens:

- public articles
- public notes
- public text that should only be stored as excerpt + summary + provenance

### Red

Do not automate by default:

- login-only surfaces
- private groups or DMs
- paywalled content
- platforms where unauthorized automated collection is likely to become a
  terms/legal problem

If something belongs here, it should only enter through manual founder drop or
an official platform connector.

## Normalized Source Unit

Every source item should eventually normalize into one shape.

Minimum fields:

- `source_id`
- `origin_type` (`founder`, `external`, `connector`, `runtime`)
- `intake_class` (`manual_drop`, `authorized_connector`, `public_collector`)
- `title`
- `url`
- `excerpt`
- `summary`
- `captured_at`
- `policy_color` (`green`, `yellow`, `red`)
- `domain_tags`
- `worldview_tags`
- `suggested_routes`

This is the point where different intake methods become one system.

## Route Proposal Layer

The system should not publish from intake directly.

It should first propose:

- keep in `97layer` only
- draft for `97layer`
- translate toward `woosunhokr`
- translate toward `woohwahae`
- archive only / no publish

Default rule:

- everything lands in `97layer` first unless the source is clearly
  destination-bound already

## Telegram Approval Layer

Telegram is the right lightweight gate for now.

Recommended actions:

- `approve 97layer`
- `approve woosunhokr`
- `approve woohwahae`
- `hold`
- `reject`
- `archive`

Current first implementation:

- founder/admin web creates the normalized source unit
- founder Telegram route can now create a raw source unit with `/drop <링크나 텍스트>`
- founder Telegram route can now use:
  - `/intake`
  - `/drafts`
  - `/route <observation_id> <97layer|woosunhokr|woohwahae|hold>`
  - `/redraft <draft_observation_id> <메모>`

That means the route decision can already move into chat.

Current automation boundary:

- `hold` stays as a stored decision only
- `97layer / woosunhokr / woohwahae` now auto-open an account-specific `draft seed`
  observation with a first body preview plus the matching proposal shell
- founder can reopen that seed with `/redraft` and keep the same
  source/account spine while writing a revised draft observation instead of
  mutating history in place
- this is still intentionally smaller than full auto-writing; it opens a real
  first paragraph without pretending the text is already finished
- founder/admin source intake now also shows those opened draft seeds inline on
  the source rows, so the route result is visible without hunting through raw
  observations
- founder/admin can now also promote a draft seed straight into a `Threads prep`
  lane, which reuses the canonical proposal/work/approval/flow publish corridor
  instead of inventing a second side path
- when that draft seed is promoted into prep, the outward prep text is now
  reshaped once more so the account voice stays visible while internal labels
  like `97layer raw draft · ...` or explicit seat-explainer paragraphs do not
  leak straight into the publish lane
- the draft seed itself is now more account-shaped too:
  - `97layer` keeps the raw operator/maker diary seat
  - `woosunhokr` leans into beauty-practice / craft translation
  - `woohwahae` leans into a quieter public shell before the actual publish lane
- if the same source is translated into more than one account, founder/admin now
  keeps those latest draft seeds side by side under the same intake row instead
  of hiding all but one

That means Telegram approval should not just say “post / do not post”.

It should also answer:

- which account
- which finish level
- whether this is a raw diary, refined craft note, or polished public note

## Translation by Account

### `97layer`

- broad raw intake
- many domains can coexist
- outside links and text can enter here first
- still needs basic tagging so the worldview does not become a pile

### `woosunhokr`

- narrower translation layer
- beauty practice
- “미용사의 단상”
- same thought, more edited and more craft-facing

### `woohwahae`

- slowest and most reduced output
- magazine-like
- public shell
- should receive translated and edited material, not raw dumps

## Immediate Build Order

1. source intake inbox
2. normalized source-unit metadata
3. Telegram route approval
4. account-specific draft opening
5. publish receipt and learning loop

The mistake to avoid is building “big crawling” before the intake shape,
approval meaning, and account routing are stable.
