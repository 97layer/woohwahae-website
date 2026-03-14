---
name: verify
description: >
  Use before handoff, release, or structural cutover to run the Layer OS
  verification lane and record the remaining risks.
---

# verify

## Purpose

Use this skill before handoff, release, or structural cutover.

## Workflow

1. run `layer-osctl verify run`
2. run contract checks
3. confirm a current preflight record exists when the change is structural or irreversible
4. record risks that remain

## Rule

Verification is a first-class lane, not an afterthought.
