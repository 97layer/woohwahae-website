# Agent Quickstart Web Design

Historical archive only.
This note is kept for continuity and should not be treated as active runtime authority or default read-path material.

## Intent

This document prepares a public website direction derived from `docs/agent-quickstart.md`.

It does not replace the current brand-home shell.
It translates worker behavior rules into public web experience rules so the site can feel like the same system.

## Why This Exists

`docs/agent-quickstart.md` is a worker prompt document, but its posture is also a useful design brief:

- practical control-tower aide, not terminal noise
- plain language first
- bounded scope
- one useful next move
- evidence after orientation, not instead of orientation

If the public website ignores those rules, the system will sound one way in prompts and feel different on screen.

## Core Translation

### From worker rule to website rule

- `orient first`
  - the first screen should explain what this surface is, why it matters, and where the visitor can go next
- `plain language first`
  - hero copy and section labels should be readable by a non-technical visitor without internal jargon
- `one lane at a time`
  - each section should have one job; avoid mixed marketing, diary, portfolio, and system status in one block
- `bounded recommendation`
  - every major section should end with one clear next action, not a menu of equally weighted choices
- `evidence supports the claim`
  - proof, routes, references, or runtime-backed signals should sit next to the main message
- `do not widen scope`
  - the home page should not try to explain the whole runtime; it should open a small number of clear doors

## Experience Goal

The website should feel:

- calm
- exact
- founder-readable
- low-temperature
- structurally confident

The website should not feel:

- promotional
- theatrical
- overloaded with categories
- decorative without evidence
- like a dashboard pretending to be a brand site

## Primary Audience

### First

- founder
- close collaborator
- operator who needs a fast orientation

### Second

- curious public visitor
- future partner who needs to understand the posture quickly

The home surface should privilege legibility over persuasion.

## Home Page Direction

### Hero job

The hero should do only four things:

1. state what Layer OS is in one short sentence
2. explain what kind of surface this is
3. show one proof-bearing side panel
4. offer one primary route forward

### Recommended hero structure

- eyebrow
  - `Layer OS public shell`
- title
  - two or three short lines only
  - example:
    - `One business shell.`
    - `Clear next moves.`
    - `Evidence before noise.`
- body
  - one paragraph that explains the value in plain language
- primary action
  - route to the main public archive or about surface
- secondary proof block
  - current posture, route map, or verified operating note

### Hero panel content pattern

The right-side panel should show:

- what is active now
- what the visitor can inspect
- what is intentionally not claimed yet

This preserves trust better than aspirational copy.

## Information Architecture

Use three public doors only at the top level.

- `Archive`
  - traces, notes, proof, and accumulated work
- `Works`
  - concrete offerings, projects, or implemented outputs
- `About`
  - founder, origin, and posture

If a fourth top-level door is proposed, it should prove why the current three cannot hold the responsibility.

## Section Blueprint

### Section 1: Orientation

Purpose:
show what the system is and how to read the site.

Content:

- short title
- one paragraph
- one proof panel

### Section 2: Operating Signals

Purpose:
show the visitor what kinds of signals matter here.

Recommended cards:

- `Clarity`
  - explain that each surface has one job
- `Evidence`
  - explain that routes, notes, and outputs back claims
- `Continuity`
  - explain that the shell is designed to stay coherent across tools

### Section 3: Public Routes

Purpose:
help the visitor choose one next path without scanning the full nav.

Recommended layout:

- 3 route cards only
- each card includes noun, short explanation, and one outcome statement

### Section 4: Founder Note

Purpose:
add a human center without turning the page into autobiography.

Content:

- one short statement from the founder posture
- one link to `About`

## Copy Rules

- prefer short nouns and verbs
- avoid visionary or inflated adjectives
- do not say `revolutionary`, `seamless`, `cutting-edge`, or similar filler
- avoid internal role names on public pages
- explain the implication before the system term
- keep paragraphs short enough to scan in one breath

### Good copy shape

- what this is
- why it matters
- what you can inspect next

### Bad copy shape

- abstract promise
- mood-heavy sentence
- unexplained system label

## Visual Direction

### Layout

- wide breathing room
- strong vertical rhythm
- asymmetric hero with proof panel
- cards used sparingly
- fewer block types across the page

### Color

- restrained light monochrome base
- near-black text
- soft grey dividers
- one muted accent only if it carries meaning

### Typography

- keep `Pretendard Variable` for body
- keep `IBM Plex Mono` for evidence labels and small structural cues
- make headlines shorter and more architectural than expressive

### Motion

- motion should reveal structure, not personality theater
- stagger only where it helps reading order
- background field remains low-contrast and reduced-motion safe
- hover and pull effects should never compete with legibility

## Component Guidance

### Header

- keep the header quiet and persistent
- top-level items should remain few
- mobile navigation should collapse cleanly without introducing a second personality

### Hero proof panel

- this is the most important non-decorative visual block
- use it to show route truth, operating posture, or verifiable notes
- avoid turning it into a badge wall

### Signal cards

- each card gets one concept only
- three cards is the default maximum for the first screen below the hero

### Footer

- keep it minimal
- sign-off should feel resolved, not busy

## Implementation Map

If this design prep moves into implementation later, the likely touch points are:

- `docs/brand-home/app/page.js`
- `docs/brand-home/components/home-shell.js`
- `docs/brand-home/components/public-surface-page.js`
- `docs/brand-home/app/globals.css`

Use those files to apply this brief.
Do not widen into runtime or contract changes for this lane.

## Non-Goals

- redesigning admin surfaces
- adding new backend contracts
- changing daemon or API behavior
- introducing a second visual identity for public pages
- turning the home page into a live runtime dashboard

## Ready-To-Build Checklist

- the first screen can be understood in under ten seconds
- the hero explains the surface before asking for attention
- each section has exactly one job
- the page offers one clear primary next move
- proof is visible near the main claim
- motion remains secondary to reading
- the page still feels calm on mobile

## Suggested Next Lane

The safest next implementation slice is:

1. tighten hero copy in `home-shell`
2. simplify the proof panel to one evidence story
3. reduce below-the-fold cards to three operating signals

That keeps the change bounded and aligned with `docs/agent-quickstart.md`.
