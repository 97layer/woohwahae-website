# Home Shell Migration

Historical archive only.
This note is kept for continuity and should not be treated as active runtime authority or default read-path material.

## Intent

Bring the legacy home shell into the current Layer OS workspace in the smallest valid unit.
The migration order is fixed:

1. backfield
2. header
3. footer
4. hero
5. inner home sections

This document is a cleaned extraction, not a verbatim port.
Legacy sources inform structure only.
Current authority for the new shell remains `constitution/` and `docs/` in this workspace.

## Baseline

### Local host

The requested visual baseline was `localhost:8080`.
At extraction time, no local server was responding on port `8080`, so the shell baseline was taken from the legacy home source and its CSS contracts.

### Source set

- `/Users/97layer/97layerOS/website/index.html`
- `/Users/97layer/97layerOS/website/_components/nav.html`
- `/Users/97layer/97layerOS/website/_components/footer.html`
- `/Users/97layer/97layerOS/website/_components/footer-logo-only.html`
- `/Users/97layer/97layerOS/website/assets/css/style.css`

## Home Shell Scope

For the first migration pass, `home shell` means only these layers:

- fixed or persistent background field
- top navigation frame
- bottom footer frame

The following are explicitly out of scope for pass one:

- hero copy
- latest article rail
- chapter rows
- site link rows
- archive/work/about content blocks
- lab-only prototypes

## Backfield

### Legacy observation

The legacy home uses a full-viewport field layer behind the hero.
Three different background ideas appear in the source history:

- `field-bg` canvas directly in home HTML
- `field-bg-canvas` fixed full-page canvas rules in CSS
- removed or disabled `wave-bg` component

The stable signal across them is not the exact rendering method but the role:

- the background is structural, not decorative
- it sits behind content at `z-index: 0`
- it is non-interactive
- it carries atmosphere with low contrast and low motion

### Current migration rule

The new home backfield should keep only the structural role:

- one fixed background layer
- monochrome only
- no ornamental overlays that compete with content
- no motion that reduces legibility
- reduced-motion safe by default

### Keep

- full-viewport background plane
- low-contrast field texture or line system
- non-interactive canvas or SVG background
- background visible mainly in the home hero zone

### Drop

- preloader dependency
- experimental cinematic overlays
- multiple stacked effect systems doing the same job
- decorative glass treatment that weakens readability

### New contract

Use one background primitive only:

- preferred: SVG or canvas field layer
- position: `fixed`
- inset: `0`
- pointer events: `none`
- default opacity: restrained
- visible purpose: depth, not spectacle

## Header

### Legacy observation

The header is a top navigation frame with:

- brand symbol link
- first-level nav links
- mobile toggle
- overlay menu
- read-progress bar

The legacy information architecture drifted between docs and code.
The code currently exposes:

- `Archive`
- `Works`
- `About`
- `Lab`

### Current migration rule

For the first migration pass, the header should carry only primary shell navigation.
It should be quiet, exact, and persistent.

### Keep

- top-aligned brand anchor
- first-level nav
- mobile toggle hook if mobile nav is needed
- sticky positioning
- minimal bottom divider

### Drop

- read-progress bar on home shell pass
- multi-column overlay menu on first pass
- secondary taxonomy in nav itself
- `Lab` as a first-class public item unless later justified

### New contract

Header v1 contains:

- brand mark or wordmark at left
- primary nav at right
- mobile toggle only when viewport requires collapse

Initial nav set for the new shell:

- `Archive`
- `Practice`
- `About`

This matches the legacy written IA more closely than the later `Works` branch.
If `Works` returns, it must be justified as a current surface noun rather than inherited by habit.

## Footer

### Legacy observation

Two footer forms exist:

- expanded footer with contact toggle, legal links, and business metadata
- home-only minimal logo footer

The expanded footer is tied to the brand-site and commerce/legal context.
The minimal footer is much closer to the current Layer OS shell posture.

### Current migration rule

For the first migration pass, home uses the minimal footer frame.
Operational or legal detail can return later only if the current shell requires it.

### Keep

- quiet footer boundary
- centered brand sign-off
- generous vertical breathing room

### Drop

- business registration block from the legacy brand site
- commerce/legal footer density in the first pass
- expandable contact panel
- live clock in footer

### New contract

Footer v1 contains:

- top border only
- centered wordmark or symbol
- no utility links by default
- no interactive control in the footer on pass one

## Visual Tokens To Carry Forward

These tokens survive conceptually from the legacy source:

- background: light monochrome paper tone
- text: near-black ink tone
- line: pale grey divider
- body font: `Pretendard Variable`
- mono font: `IBM Plex Mono`
- spacing scale biased toward large gaps
- single visual family across shell parts

These legacy implementation details do not carry forward automatically:

- exact token values
- glassmorphism layer stack
- cinematic naming
- preloader choreography
- Three.js dependency as a requirement

## Migration Decisions

### Pass One Build

The first implementation unit is:

- `home shell`
  - `backfield`
  - `header`
  - `footer`

### Pass Two Build

Only after pass one settles:

- `hero`
- `home intro row`
- `home site nav row`
- `home featured row`

### Naming Direction

Use current Layer OS naming, not legacy performance naming.
Prefer:

- `home-shell`
- `shell-header`
- `shell-footer`
- `home-backfield`

Avoid:

- cinematic or theatrical class names unless the visual role truly requires them
- restoring legacy specialist vocabulary

## Done Condition For Pass One

Pass one is done when the current workspace has a home shell that satisfies all of the following:

- one fixed monochrome backfield exists
- one sticky header exists
- one minimal footer exists
- the three pieces read as one shell
- there is no unnecessary visual effect stack
- the shell remains legible before any hero content is added
