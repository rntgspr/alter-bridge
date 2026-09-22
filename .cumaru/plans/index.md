---
human_revised: false
targets: [meta]
summary: Framework guidance for Plans and its required workflow.
---


# Plans

A transient pillar for **active work items**. One directory per tracker-backed
item or maintenance initiative. The Lead authors `index.md` and `t<N>.md` at
planning time; on close the Lead absorbs the delta into `specs/` and removes
the complete plan after the acceptance, validation, and recovery gates pass.

## Rules

- **One directory per plan.** `plans/<KEY>/` for tracker-backed work, `plans/maintenance-<slug>/` for internal initiatives. The directory name is the plan ID.
- **Slug-based plans require the `maintenance-` prefix.** Pure kebab-case slug (`maintenance-cleanup-deprecated-helpers`). No `key:` frontmatter field.
- **Plan body carries `## Overview`, `## Acceptance Criteria (EARS / RFC 2119)`, `## Plan / DAG`, `## Out of scope`, `## Risks`.** Unlike SDLC, there is no separate intake pillar — everything lives in the plan body.
- **Tasks within a plan may run in parallel** when `depends-on:` is satisfied and `files:` predictions do not overlap.
- **Each entry is a directory** with `index.md`, `t<N>.md` per task, optional `handoff-t<N>.md`, and `delta-draft.md` at close.

## When to use

- Starting work on a tracker item or internal initiative → create `plans/<PLAN-ID>/` (Lead).
- Implementing tasks of an active plan → flip `t<N>.md` `status:` and write `handoff-t<N>.md`.
- Closing a plan → Lead validates acceptance and the delta draft, establishes
  the required recovery point, absorbs into specs, then removes the plan.

## When NOT to use

- Description of the system as it is today → `specs/<area>/`.
- Pre-plan ideation, sketches, options analysis → `exploring/<slug>/`.

## Go rewrite implementation order

Adopter-owned sequencing note for the bridge's bash-to-Go port. Each entry
is a `plans/maintenance-go-<command>/` scaffold; work them top to bottom
unless a dependency is explicitly relaxed.

<!-- cumaru:go-rewrite-order -->
1. `maintenance-go-send` — addressing, message format, atomic write. Everything below reuses this.
2. `maintenance-go-inbox` — drain + archive; the core read path.
3. `maintenance-go-peek` — shares `inbox`'s read helper, minus archiving.
4. `maintenance-go-who` — independent (session-state lookups), low risk, good to interleave early.
5. `maintenance-go-archive` — reuses `inbox`'s archive-move logic for a whole mailbox or every mailbox.
6. `maintenance-go-purge` — deletes what `archive` produced.
7. `maintenance-go-hook` — wraps `send`'s addressing + `inbox`'s drain as the UserPromptSubmit entry point.
8. `maintenance-go-watchpaths` — SessionStart registration; the doorbell's first half.
9. `maintenance-go-relay` — FileChanged reaction to what `watchpaths` registered; most complex, land last.
<!-- /cumaru:go-rewrite-order -->
