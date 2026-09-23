---
human_revised: false
plan: maintenance-go-relay
task: T3
status: blocked
date: 2026-09-23
summary: Hand-off for maintenance-go-relay T3 — Blocked; AC8 live verification must be run by the user with the steps in t3.md, since agents may not spawn a real claude -p or wake a real session.
---

# Hand-off — maintenance-go-relay / T3

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
<!-- /cumaru:touched -->

## Decisions made during implementation

- Not executed by the agent: the safety rules forbid spawning a real `claude -p`, waking a real session, or deleting a real transcript. The steps live in `t3.md` `## Implementation` and isolate the probe under `~/.alter-bridge-probe` so the global bash hooks stay silent.

## Commands run / verification

- None. AC8 is **Blocked (needs user-run live verification)**.

## Pending / follow-ups

- User runs `t3.md` steps 1-8 and records steps 5-7 outcomes here; flip `t3.md` and this handoff to done on PASS, then close the plan with `cumaru-absorb` using `delta-draft.md`.

## Suggestions for the Lead

- Do not absorb or prune this plan until AC8 is PASS.
