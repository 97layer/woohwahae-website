---
name: sequence
description: >
  Use when the task needs explicit step-by-step reasoning, option comparison,
  or staged decision framing before or during implementation.
---

# sequence

## Purpose

Use this skill when the task needs explicit step-by-step reasoning, option comparison, or staged decision framing.

This is the local Layer OS replacement for depending on an external Sequential Thinking tool.

## When To Use

- architecture tradeoff decisions
- feature prioritization
- migration planning
- bug isolation with multiple hypotheses
- any task where the user asks for sequential thinking explicitly

## Workflow

1. restate the goal in one sentence
2. list current facts from repository contracts, docs, and runtime surfaces
3. name the unknowns or risks
4. generate 2-4 concrete options
5. compare options by value, risk, and implementation cost
6. choose one recommendation
7. if implementation follows, execute the first concrete step immediately

## Output Shape

Prefer this response order:

1. `Goal`
2. `Facts`
3. `Options`
4. `Recommendation`
5. `Next Step`

Keep it compact.
Do not invent hidden certainty.
Use repository evidence before intuition.

## Rule

- contracts before speculation
- architecture before patching
- recommendation must include why not the rejected options
- if the user asks for execution, stop after recommendation only long enough to validate the write target, then implement
