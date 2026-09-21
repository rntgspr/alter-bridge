# Project instructions

<!-- BEGIN CUMARU-HOOK created -->
## `.cumaru/` framework

This project uses the `.cumaru/` framework — a spec-driven, agent-friendly knowledge structure. At the start of every session in this repository, enter the framework in this order: read `.cumaru/index.md` (the kernel), `.cumaru/domain.md` (the domain), `.cumaru/disciplines/index.md` (the discipline evaluation contract), and every other installed file under `.cumaru/disciplines/`, then run `cumaru tree .` to project the root's current candidates and their summaries. Every discipline is loaded. Its `strictness` controls required consideration and its `applies-when` controls application; a missing strictness is invalid and treated as `0/10`. Prune tree candidates by relevance — never prune the execution disciplines.

@.cumaru/index.md
@.cumaru/domain.md
@.cumaru/disciplines/index.md
@.cumaru/disciplines/acceptance-testing.md
@.cumaru/disciplines/blast-radius.md
@.cumaru/disciplines/code-comments.md
@.cumaru/disciplines/cumaru-first.md
@.cumaru/disciplines/dry.md
@.cumaru/disciplines/engineering.md
@.cumaru/disciplines/kiss.md
@.cumaru/disciplines/receiving-code-review.md
@.cumaru/disciplines/solid.md
@.cumaru/disciplines/systematic-debugging.md
@.cumaru/disciplines/test-driven-development.md
@.cumaru/disciplines/verification.md
@.cumaru/disciplines/yagni.md
<!-- END CUMARU-HOOK -->
