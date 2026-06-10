# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A single-file static website for Anne Nguyen and David Zech's wedding (June 5, 2027). The `index.html` is a saved export of their The Knot wedding page — a large, self-contained HTML file with inlined styles and embedded JSON data from The Knot's Next.js frontend.

## Structure

The entire site is `index.html` (~2,500 lines). It contains:
- `<head>` — Next.js-generated meta tags, preloaded CSS/font links referencing `static.theknot.com`
- Inline `<script id="__NEXT_DATA__">` — a large JSON blob with all wedding content (venue, schedule, registry, photos, etc.)
- Rendered HTML — the visible page content as server-rendered by The Knot

There is no build step, no package manager, and no server. Open `index.html` directly in a browser to view it.

## Editing

All wedding content lives inside the `__NEXT_DATA__` JSON blob and the rendered HTML below it. External assets (images, CSS, fonts) are still served from The Knot's CDN — the page requires an internet connection to display correctly.

To find specific content (venue address, RSVP details, registry links, event schedule), grep for keywords inside `index.html`.
