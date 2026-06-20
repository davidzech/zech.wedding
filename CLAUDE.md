# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A single-purpose Go reverse proxy. It serves `zech.wedding` by proxying every
request to one fixed The Knot wedding page (`target` in `main.go`) and patching
the HTML on the way through so the couple can customize a page they don't
control. There are no external Go dependencies — standard library only.

## Commands

```bash
go build -o /tmp/zechproxy .   # build (Homebrew Go may be at /opt/homebrew/bin/go)
go run .                       # run locally on :8080 (override with PORT=xxxx)
go vet ./...                   # vet
```

There are no tests. To verify a change end-to-end, run the proxy and curl it:

```bash
go run . &
curl -s http://localhost:8080/ | grep -c "<thing you changed>"
```

Note: client-side patches (see below) will **not** appear in `curl` output —
they only run in a real browser once React executes. Only server-side rewrites
are visible via curl.

## The two patching mechanisms — and how to choose

The Knot page is a React/Next.js app. Content is either in the initial HTML
response (server-rendered) or injected into the DOM later during hydration /
client-side navigation. **Which mechanism to use depends on which it is.**

Always determine this first by fetching the page and grepping for the markup
(check both the plain and the JSON-escaped `<` form, since Next embeds a
serialized copy of the DOM in its hydration payload):

```bash
curl -s --compressed "<target URL>" -o /tmp/knot.html
grep -c '<thing>'           /tmp/knot.html   # plain / server-rendered
grep -c 'u003c<thing>'      /tmp/knot.html   # escaped payload copy
```

1. **Server-side rewrite** (`main.go`, `ModifyResponse`) — use when the content
   is in the HTML response. The handler decompresses gzip, does `bytes.Replace`
   on the body, then re-sets `Content-Length`. Best when something must be
   neutralized *before* the browser ever executes it. Example: the OneTrust
   cookie popup is killed here by rewriting its `otSDKStub.js` script URL to an
   empty data URI — the inline loader still runs and `window.UnionConsentManagement`
   stays defined (other inline scripts call it unguarded), but the SDK never
   loads, so the popup is never built.

2. **Client-side DOM patch** (`patch.html`) — use when the content is rendered
   by React at runtime (not in the HTML). This file is embedded via `//go:embed`
   and injected before `</body>` on every HTML response. All DOM edits live in
   `applyPatches()`, which is driven by a `MutationObserver` so it reapplies as
   React mutates the DOM, including across client-side anchor-link navigations.

## Conventions for client-side patches in `patch.html`

- **Target stable selectors.** The `wws-*` CSS classes are hashed per build and
  change over time — prefer `data-testid` or visible text content. When you must
  use a `wws-*` class, expect it to break eventually and document it.
- **Make every patch idempotent.** `applyPatches()` runs on every batch of DOM
  mutations. Patches that add nodes must guard against re-adding (e.g. a
  `dataset.*` flag, or checking the target is non-empty). Patches that remove
  nodes are naturally idempotent (the match disappears).
- **The observer stays connected for the page's lifetime** and is debounced to
  one run per `requestAnimationFrame`. Do not re-add a `disconnect()` on load —
  that breaks reapplication after client-side navigation.
- **Prefer a CSS rule over JS** for styling changes: a `<style>` block in
  `patch.html` applies to current and future elements automatically without
  needing to run in `applyPatches()`. Use `!important` to beat the site's
  CSS-in-JS, which injects styles dynamically.

## After editing `patch.html`

It is compiled into the binary via `//go:embed`, so **rebuild** (and restart any
running instance) for changes to take effect.

## Other proxy behavior worth knowing

- `Rewrite` (not the deprecated `Director`) pins host/path/query to `target` and
  forces `Accept-Encoding: gzip`; `ModifyResponse` decompresses it back.
- `Content-Security-Policy` and `X-Frame-Options` are stripped so the injected
  inline script/style can run.
- Listens on `$PORT` (default `8080`) — App Platform injects this.

## Deployment

DigitalOcean App Platform, configured by `.do/app.yaml` (builds from the
`Dockerfile`, deploys on push to `main`, binds the `zech.wedding` apex domain,
TLS auto-provisioned). The `distroless/static` runtime image includes CA certs,
which the proxy needs for its outbound HTTPS call to The Knot.
