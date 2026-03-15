# my-personal-site (rename to match your repo)

## Overview

* **Markdown authoring pipeline** — write content in `.md`, generate clean HTML via pandoc using templates you fully control. No framework opinions, no dependencies to keep current — `make` targets wrap the whole pipeline into single commands. The `.md` source files live alongside the HTML and are served directly, making content accessible to crawlers and AI indexers without extra work.

* **RSS feed** — `feed.xml` is auto-generated and served at `/feed/` with the correct `Content-Type`. RSS lets readers follow your writing on their own terms — no social media account required on either side, no algorithm between you and your audience. Adding a `description` field to an article's frontmatter populates the feed item summary.

* **SVG favicons with `.ico` fallback** — modern browsers use `favicon.svg`, which scales cleanly at any size and can embed a `@media (prefers-color-scheme)` block to adapt to dark mode. Safari doesn't support SVG favicons, so a `favicon.ico` (multi-resolution: 32px + 16px) is included as a fallback. Favicon colors are automatically kept in sync with each site's CSS custom properties — change a color token in `style.css` and the favicon updates on the next build.

* **Custom 404 and 503 error pages** — a generic browser 404 is a dead end that makes your site feel unfinished. The worker serves a branded 404 page that keeps the visitor in context. The 503 page is shown for requests to an unknown domain, preventing a blank response in misconfiguration scenarios. Both are hand-coded HTML outside the build pipeline.

* **Two or more domains, one deployment** — the worker routes by hostname, so two distinct sites can share one repo and one Cloudflare Worker. Useful for keeping a professional/CV site and a writing/projects site (plus any others you want to add) separate in identity while managing them together.

* **Local development with multi-device preview** — `make -C maint serve` starts wrangler bound to `0.0.0.0`, making the local server reachable from any device on your network. Check mobile layout and Safari rendering without deploying.

## Cheat Sheet
Day-to-day authoring reference for this two-domain Cloudflare Worker site. The build pipeline converts Markdown to HTML via pandoc, driven by site Makefiles in each content root (`public/[site]/Makefile`). The authoring utilities at a glance:

* **New article** — `cd public/my-personal-site.me && make new-article SLUG=my-slug` scaffolds `writing/my-slug/index.md` from the article template. Edit the `.md` file, then publish with `cd public/my-personal-site.me && make`.
* **Build and publish** — `cd public/<domain-name.tld> && make` generates all HTML, updates the article index and RSS feed, and syncs favicon colors.
* **Favicon colors** — automatically synced from each site's `style.css` when running `make`. Sync manually with `make sync-svg-colors` from the site directory.
* **site-tool** — the Go binary underlying the pipeline: `gen-toc` (article index + RSS), `sync-svg-colors` (CSS → SVG color sync), `glyphs` (font path extraction for favicon design). See the site-tool section below.
  * You generally will not be using site-tool directly. It's invoked by `maint/gen-html.sh` and the site Makefiles.

## Domains

| Domain | Content root |
|---|---|
| `my-personal-site.me` | `public/my-personal-site.me/` |
| `my-second-personal-site.me` | `public/my-second-personal-site.me/` |

The Worker (`src/index.js`) routes by hostname — each domain maps directly to its directory under `public/`. Unknown domains return a custom 503 error page.

## Adding or removing domains

**Adding a domain** — copy an existing content root directory (e.g. `cp -r public/my-personal-site.me public/new-site.com`), then edit the content to fit the new site. The Makefile rules use relative paths with no hardcoded domain variables, so nothing in the build config needs changing. `src/index.js` is updated automatically on the next `make` invocation from either site directory.

**Removing a domain** — delete the content root directory, then run:

```sh
make -C maint sync-domain-names
```

This updates `src/index.js` to reflect the remaining domains.

## Local development

```sh
make -C maint serve
```
NOTE: this is synchronous.
Wrangler prints your LAN IP in its startup banner.
You can use `http://<your-lan-ip>:8787/` from any device on the network.

Locally there is no hostname distinction, so use the `?d=` query parameter to select a site on first load:

- `http://localhost:8787/?d=my-second-personal-site.me` → my-second-personal-site.me
- `http://localhost:8787/?d=my-personal-site.me` → my-personal-site.me

This sets a session cookie, so subsequent internal navigation works without repeating the parameter. Also works on `*.workers.dev` preview URLs.

NOTE: cross-site (my-personal-site.me <-> my-second-personal-site.me) links won't work in local development because they use absolute urls (those will always point to the deployed site, not the local work in progress)

NOTE: wrangler monitors changes to all files, including `src/index.js` - you don't need to kill and restart it to observe edits.

## Content authoring

All site HTML files besides the 404 and 503 error pages are generated from paired `.md` sources via pandoc. Site targets run from each site's directory; infra targets (`serve`) run from `maint/`:

```sh
# my-personal-site.me
cd public/my-personal-site.me && make              # build everything + sync SVG colors

# my-second-personal-site.me
cd public/my-second-personal-site.me && make        # build everything + sync SVG colors
```

There are no separate per-page targets — `make` (i.e. `make public`) builds everything.

Use `make -B` to force a full rebuild when only a filter changed — Lua filters aren't tracked in Make's dependency graph, so changes to them won't trigger a rebuild automatically.

## Adding pages

**New writing article** — `cd public/my-personal-site.me && make new-article SLUG=my-slug` scaffolds `writing/my-slug/index.md` from the article template. The new page is picked up automatically on the next `make`.

**New generated page** (any `.md → .html` page) — create the directory and drop in `index.md` plus a sibling `index.template.html`. Make's recursive file scan picks it up automatically — no Makefile changes needed. Specify any Lua filters in the frontmatter `filters:` list.

### Adding sets of pages with a shared template
* Create a subdirectory that holds all of the page directories that will share the template (e.g. `the/group/`)
* Put the shared template in the group directory (e.g. `the/group/shared.template.html`)
* Add a rule to the content root Makefile for handling it as a special case
```makefile
# .md -> .html override rule for pages that share a common template instead of a sibling .template.html per .md file
# $(<) expands to the name of the first dependency (the markdown file)
# $(@) expands to the name of the current target (the .html file being built)
the/group/%/index.html: the/group/%/index.md the/group/shared.template.html
	MD=$(<) TEMPLATE=the/group/shared.template.html HTML=$(@) $(GEN_HTML_SH)
```
* Create the pages as usual under the group directory (e.g. `the/group/a-page/index.md`, `the/group/b-page/index.md`)
* `make` will build the HTML pages using the shared template

**Hand-authored HTML page** (no `.md`) — create the directory and drop in `index.html`. Nothing else needed.

## RSS feed

`public/my-personal-site.me/feed.xml` is generated automatically by `site-tool gen-toc` as part of `make`. It is served at `/feed/` (and `/feed`) with the correct `Content-Type`.

To include an item description in the feed, add a `description` field to the article's YAML frontmatter.

## Favicons

Both sites use `favicon.svg` + `favicon.ico` (Safari fallback). The SVG files use `<path>` elements extracted from actual font files via `site-tool glyphs` for consistent cross-platform rendering — do not replace paths with `<text>` elements.

SVG colors are kept in sync with each site's `style.css` via `site-tool sync-svg-colors`. SVG class names map directly to CSS custom property names (e.g. class `accent` → `--accent`). Run automatically as part of `make`, or manually:

```sh
cd public/my-personal-site.me && make sync-svg-colors
cd public/my-second-personal-site.me && make sync-svg-colors
```

To regenerate an `.ico` from its SVG (requires ImageMagick):

```sh
cd public/my-personal-site.me && make favicon
cd public/my-second-personal-site.me && make favicon
```

On macOS, qlmanage is used for reliable SVG rendering. On Windows and Linux, ImageMagick's SVG renderer is used directly — verify the output looks correct. If the result is bad, fall back to a free online SVG→ICO converter as a one-off. This is rarely needed — only when the favicon design changes.

## site-tool

`maint/site-tool.sh` is a platform-detecting wrapper around a compiled Go binary in `maint/bin/`. Pre-compiled binaries for all supported platforms are checked in via Git LFS. To build from source:

```sh
make -C maint site-tool        # build for current platform
make -C maint site-tool-dist   # cross-compile all platforms
```

Subcommands: `glyphs`, `gen-toc`, `sync-svg-colors`, `version`. Run `site-tool <command> -help` for options.

Read AGENTS.md for build pipeline details, author contracts, and content conventions.
