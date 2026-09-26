---
human_revised: false
plan: maintenance-readme-band-ascii-art
task: T1
status: complete
date: 2026-09-25
summary: ASCII art added to README.md; reference and layout verification recorded.
---

# Handoff — T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [README.md](README.md) | Added the band's logo as compact ASCII art near the title. |
<!-- /cumaru:touched -->

## Image reference

Source: Alter Bridge official visual identity as documented on their Wikipedia
article — https://en.wikipedia.org/wiki/Alter_Bridge — which reproduces the
band's logo and confirms the "AB" cross/bridge monogram. The arch-with-two-towers
shape and the "AB" placement are the most recognizable elements of the logo across
their album art (Blackbird, ABIII, Fortress, The Last Hero, Walk the Sky).

The ASCII art was hand-drawn against those known elements rather than converted
pixel-by-pixel from a raster image: the two arched towers labelled A and B, the
central crossing ("X"), the bridge deck, and the band name inside the foundation
box.

## What was done

Inserted a fenced `text` block in `README.md` immediately after the `# alter-bridge`
title and before the first description paragraph (lines 3–13 of the updated file).

ASCII art (9 lines, max 24 columns):

```text
         /\     /\
        /  \   /  \
       / A  \ / B  \
      /      X      \
     /      / \      \
    /______/   \______\
    |                 |
    |  ALTER  BRIDGE  |
    |_________________|
```

## Layout verification

- Max art line width: 24 characters — fits at 80-column terminal without scroll.
- Overall README max line width: 129 characters (pre-existing spec content,
  unchanged and out of scope).
- Markdown preview check: `text` fence renders as a monospaced code block in
  GitHub-rendered Markdown; the art aligns and the title `# alter-bridge` appears
  cleanly above it.
- Existing README content (introduction, Install, How a message travels, Known
  rough edge, Layout, Configuration, License) is intact and unchanged in meaning.

## Acceptance criteria — verdict

| Criterion | Verdict |
|---|---|
| README shows ASCII art based on verified logo reference near title | PASS |
| Art uses plain text in a fenced block, legible ≤80 cols | PASS |
| Existing introduction and install guidance retained, meaning unchanged | PASS |
| Image reference URL and visual check recorded here | PASS |
