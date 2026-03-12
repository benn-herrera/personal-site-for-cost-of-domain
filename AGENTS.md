# AGENTS.md — Technical Context for AI Assistants

This file is for AI coding assistants helping users through the process described in `ZERO_TO_LAUNCH.md`. It provides technical depth about the codebase so you can give accurate, specific guidance without the user needing to explain what the code does.

## Project Purpose

This is a template repo for hosting one or more personal websites on Cloudflare Workers. The cost to run it is exactly the cost of the domain — Cloudflare Workers and GitHub are both free at this scale. The worker serves static HTML/CSS/JS files from `public/` and routes requests to the right subdirectory based on the incoming domain name.

## File Map

```
wrangler.jsonc                        — Cloudflare Workers config
src/index.js                          — The worker: all request routing logic lives here
public/
  my-personal-site.me/
    index.html                        — Placeholder for site 1
  my-other-personal-site.me/
    index.html                        — Placeholder for site 2 (delete if not needed)
```

## The Critical Coupling

**The `sites` array in `src/index.js` must exactly match the directory names under `public/`.** This is a potential source of confusion. When a user renames directories or adds a new domain, they must update both places.

```js
// src/index.js line 4 — must mirror public/ subdirectory names
const sites = ["my-personal-site.me", "my-other-personal-site.me"];
```

If a domain isn't in `sites`, the worker returns 503. If a directory doesn't exist in `public/` for a domain that is in `sites`, asset fetches will 404.

## How the Routing Works

The worker intercepts every request (`run_worker_first: true` in `wrangler.jsonc`). It determines which `public/<domain>/` subtree to serve from based on the incoming hostname:

**Production (real domain):** `url.hostname` is the actual domain (e.g. `my-personal-site.me`). Used directly to pick the content root. No cookie or `?d=` logic involved.

**Dev / preview (localhost, 127.x.x.x, 192.x.x.x, *.workers.dev):** The worker can't use the hostname to pick a site. Instead it uses a `?d=<domain>` query parameter. Once set, `?d=` writes a cookie so internal links and navigation within the site continue to work. On subsequent requests without `?d=`, the cookie value is used. Falls back to `sites[0]` if neither is present.

Asset URL construction (line 43):
```js
const assetUrl = new URL(`/${domain}${pathname}`, url.origin);
```
So a request for `/about.html` on `my-personal-site.me` fetches `public/my-personal-site.me/about.html`.

**RSS alias:** `/feed` and `/feed/` are rewritten to `/feed.xml` before asset lookup (lines 38–40).

## Key Config: wrangler.jsonc

- `"name"` — must match the Worker name chosen on Cloudflare. This also determines the default `workers.dev` URL: `<name>.<cf-account>.workers.dev`. Keep it in sync with the repo name to avoid confusion.
- `"assets": { "directory": "./public", "binding": "ASSETS", "run_worker_first": true }` — `run_worker_first: true` is what makes the worker code run before Cloudflare's default asset serving. Without it, static files would be served directly and the routing logic would never execute.
- `"compatibility_date"` — controls which Cloudflare runtime APIs are available. Don't change this unless you know why.

## Common User Tasks and What to Change

**Rename for a single domain:**
1. Rename `public/my-personal-site.me/` to `public/<their-domain>/`
2. Delete `public/my-other-personal-site.me/` (or rename it too)
3. Update the `sites` array in `src/index.js` to `["<their-domain>"]`
4. Update `wrangler.jsonc` `"name"` field to match their worker project name

**Add a domain:**
1. Create `public/<new-domain>/` with at least an `index.html`
2. Add `"<new-domain>"` to the `sites` array in `src/index.js`

**Remove a domain:**
1. Remove the entry from `sites` in `src/index.js`
2. Optionally delete the `public/<domain>/` directory

**Verify changes locally:**
Stop `wrangler dev` (press `q`), restart it, then test with `?d=<domain>` in the URL.

## Local Dev URL Tricks

- `http://localhost:8787` — serves `sites[0]` by default
- `http://localhost:8787?d=my-personal-site.me` — switches to that site and sets cookie
- `http://192.168.x.x:8787` — same routing rules; accessible from other devices on local network (requires `wrangler dev --ip 0.0.0.0`)

The `?d=` cookie persists for the browser session. To switch sites, use `?d=` again with the target domain. To reset to default, clear cookies or open a fresh browser profile.

## Wrangler Commands

```sh
wrangler dev --ip 0.0.0.0   # local dev server, available on LAN
wrangler login               # authenticate with Cloudflare (one-time, opens browser)
wrangler deploy              # manual deploy from local files (bypasses git)
```

`wrangler deploy` deploys whatever is on disk right now — it has no awareness of git state. Warn users to be on a clean main branch before running it.

## Deployment Behavior

- Push to `main` on GitHub → automatic deploy (usually ~30 seconds)
- Push any other branch → preview deploy at `https://<branch-name>-<worker-name>.<cf-account>.workers.dev`
- Branch previews use the same `?d=` routing for multi-site testing
- Auto-deploy can be disabled in Worker settings on Cloudflare dashboard
- If auto-deploy gets stuck, `wrangler deploy` forces it

## Adding Site Content

Everything under `public/<domain>/` is served as static assets. Subdirectories work as expected (`/blog/post.html` → `public/<domain>/blog/post.html`). There is no build step — the files are served as-is by Cloudflare's asset pipeline.

## Cloudflare Setup Sequence (Technical Notes)

When creating the Worker on Cloudflare:
- Choose **Worker** (not Pages) — this project uses the Worker + Assets pattern, not Pages
- Connect GitHub repo and limit access to this repo only
- The worker name on Cloudflare must match `"name"` in `wrangler.jsonc`

When adding a custom domain to the worker:
- Add both the bare domain (`example.me`) and the www alias (`www.example.me`) as custom domains on the Worker's Settings page
- Cloudflare handles the DNS automatically when the domain is registered through Cloudflare
- For external registrars, NS records must point to Cloudflare before custom domains will work
