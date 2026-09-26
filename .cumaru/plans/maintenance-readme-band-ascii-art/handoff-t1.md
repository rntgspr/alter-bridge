---
human_revised: false
plan: maintenance-readme-band-ascii-art
task: T1
status: complete
date: 2026-09-25
summary: Revised the README with a 76-column arched ASCII wordmark based on the official band logo, with local browser and content-preservation checks.
---

# Handoff - T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [README.md](README.md) | Replaced the small bridge sketch with a bold, arched ASCII wordmark. |
<!-- /cumaru:touched -->

## Image reference

- Source: [Alter Bridge official website](https://alterbridge.com/).
- Direct reference: [official header wordmark](https://alterbridge.com/cdn/shop/files/AB_Web_Logo_1360x.png?v=1767913113).
- Retrieved and visually inspected on 2026-09-25. The reference uses heavy,
  condensed capitals with an arched silhouette. The ASCII adaptation preserves
  those features with explicit gaps between letters and between the two words.

The earlier handoff's description of two towers and an AB crossing was not
supported by an inspected image. This revision replaces that unsupported
description and its associated verification claims with the reference above.

## What was done

Replaced the opening fenced text block with a 12-line, 76-column ASCII
wordmark. The lettering is authored for monospace readability, not a
pixel-exact reproduction. No image asset or rendering dependency was added.
The canonical artwork is in README.md; it is not duplicated here.

## Layout verification

Executed `node /tmp/alter-bridge-preview.cjs` against the revised README:

- ASCII characters only; 12 rows; maximum width 76 columns; no trailing spaces.
- Replacing the opening text fence with an empty string in both HEAD and the
  working copy produces identical content. All other README text is unchanged.
- A local Chromium fixture rendered the actual opening text block with
  GitHub-like monospace styling at 1280 px and 390 px viewport widths.
- Desktop: the entire wordmark fits without horizontal scrolling.
- Mobile: the page fits the viewport; the code block scrolls horizontally,
  preserving letter alignment. The full wordmark is not visible at once.
- Both screenshots were visually inspected. Temporary evidence paths:
  `/tmp/alter-bridge-preview-1280.png` and
  `/tmp/alter-bridge-preview-390.png`.

The browser check covers the opening in a local fixture, not GitHub's live
Markdown renderer. No CLI tests or build were run for this documentation-only
change.

## Acceptance criteria - verdict

| Criterion | Verdict | Evidence |
|-----------|---------|----------|
| ASCII art based on a verified band logo near the title | PASS | Official wordmark downloaded and visually inspected; revised opening checked. |
| Plain text in a fenced block, legible within 80 columns | PASS | ASCII assertion and measured maximum of 76 columns; desktop screenshot inspected. |
| Existing introduction and installation guidance retained | PASS | Exact comparison outside the opening text fence against HEAD. |
| Reference URL and rendered/plain-text layout checks recorded | PASS | Direct official image URL, terminal output, and local Chromium checks above. |
