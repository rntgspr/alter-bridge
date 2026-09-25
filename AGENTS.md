# Project instructions

<!-- BEGIN CUMARU-HOOK created -->
## `.cumaru/` framework

This project uses the `.cumaru/` framework — a spec-driven, agent-friendly knowledge structure. At the start of every session in this repository, enter the framework in this order: read `.cumaru/index.md` (the kernel), `.cumaru/domain.md` (the domain), `.cumaru/disciplines/index.md` (the discipline evaluation contract), and every other installed file under `.cumaru/disciplines/`, then run `cumaru tree .` to project the root's current candidates and their summaries. Every discipline is loaded. Its `strictness` controls required consideration and its `applies-when` controls application; a missing strictness is invalid and treated as `0/10`. Prune tree candidates by relevance — never prune the execution disciplines.

@.cumaru/index.md
@.cumaru/domain.md
<!-- END CUMARU-HOOK -->
