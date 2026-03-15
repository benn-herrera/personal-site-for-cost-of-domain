# Agent Guide

Cloudflare Worker project backing two domains from one repo.

## domains
* my-second-personal-site.me — professional identity, CV, materials for hiring teams
* my-personal-site.me — writing, tech/process articles, project links

## philosophy
* content: targeted, purposeful
* aesthetics: lean and restrained — simplicity is a deliberate statement
* no frameworks
  * strong preference for hand-coded HTML/CSS only
  * resort to JS only if a necessary behavior would be ugly/awkward without it
* pages set up for crawler discoverability: parallel .md files alongside most HTML articles and the CV

## worker technical setup
* Worker config: wrangler.jsonc
* public/ — top-level content root
  * public/my-second-personal-site.me/ — my-second-personal-site.me content root
    * public/my-second-personal-site.me/style.css — canonical palette and aesthetic reference for this site
  * public/my-personal-site.me/ — my-personal-site.me content root
    * public/my-personal-site.me/style.css — canonical palette and aesthetic reference for this site
* HTTP request handling: src/index.js
  * routes by domain: hostname → public/{domain}/; unknown domain → 503 (no fallback)
  * `const sites = [...]` in src/index.js must match subdirectory names under public/ — hardcoded by necessity since the Worker runtime has no filesystem access to discover them dynamically
  * local dev: ?d=[site-domain.tld] sets a session cookie so subsequent internal links inherit the domain without the param; activates for localhost, 127.x.x.x, 192.x.x.x, *.workers.dev
  * cross-site links are absolute URLs so will always point to the deployed site, not the local work in progress.
  * `/feed` and `/feed/` — URL alias → `public/[site-domain.tld]/feed.xml` with `Content-Type: application/rss+xml`

## content conventions
* AI-generated prose is prefixed "AI GEN PLACEHOLDER:" so the author can identify what needs rewriting
* articles and story blocks are reverse chronological
* dev articles are authored in .md; HTML is generated from them (see source files table)
* nav: aria-current="page" on the active nav destination; omitted on article pages (articles are not nav items)
* topic tags (.tag class) are purely decorative — no hover behavior, no filtering
* cross-site links: .dev→.me use "at my-second-personal-site.me ↗" in link text; .me→.dev use "at my-personal-site.me ↗"

---

## build pipeline (maint/)

HTML files that have a paired `.md` source are generated from that `.md` via pandoc. The expected workflow is to `cd` into a site's content root and run `make` targets from there:

```sh
cd public/my-personal-site.me && make public
cd public/my-second-personal-site.me && make public
```

Infra targets (`serve`, `site-tool`) are in `maint/Makefile` and are run from `maint/`:
```sh
cd maint && make serve
```

### Makefile structure

Each site content root (`public/[site]/`) has a `Makefile` that includes two shared files from `maint/`:

- **`maint/common.mk`** — included first; provides `DOMAIN` (auto-derived from directory name), `SITE_AUTHOR`, `SITE_TOOL_SH`, `GEN_HTML_SH`, recursive `MD_SRCS`/`HTML_TARGETS` via `find`, and the base `%.html: %.md %.template.html` pattern rule.
- **`maint/favicon.mk`** — included after the site's default target; provides `sync-svg-colors` and `favicon` targets.

Site Makefiles define only what's site-specific on top of these shared rules.

### Makefile targets
| target | what it does | where |
|---|---|---|
| `public` | build all HTML + sync SVG colors + favicon (default target) | each site's `Makefile` |
| `sync-svg-colors` | sync CSS colors into favicon.svg (UTD checked against style.css) | `maint/favicon.mk` |
| `favicon` | regenerate favicon.ico from favicon.svg (requires ImageMagick; macOS uses qlmanage, Windows/Linux use ImageMagick directly — verify output) | `maint/favicon.mk` |
| `new-article` | scaffold a new article at `writing/$SLUG/index.md` from template (requires `SLUG=name`) | `public/my-personal-site.me/Makefile` |
| `serve` | start wrangler dev on 0.0.0.0 (wrangler prints LAN IP in its startup banner) | `maint/Makefile` |

Use `make -B [target]` to force a rebuild when only a filter changed — those are not in Make's dependency graph.

### pipeline components
```
maint/
├── Makefile                  ← infra targets only: serve, site-tool
├── common.mk                 ← shared config, MD_SRCS/HTML_TARGETS, base %.html rule
├── favicon.mk                ← shared sync-svg-colors and favicon targets
├── gen-html.sh               ← pandoc wrapper: parses filters: from frontmatter, invokes pandoc
├── site-tool.sh              ← platform-detecting wrapper; invokes bin/site-tool or bin/site-tool-{os}-{arch}
├── bin/                      ← compiled site-tool binaries (gitignored dev build + LFS dist builds)
├── site-tool/                ← Go source for site-tool binary
└── filters/
    ├── footnotes.lua         ← custom footnote links for writing articles
    └── foldout.lua           ← foldout/collapsible section support
```

### .md → .html pipeline

The base rule in `common.mk` handles all `.md` → `.html` conversions. Each `.md` file expects a sibling `.template.html`. For writing articles, a directory-level shared template (`writing/article.template.html`) is used via an override rule in the site Makefile. `gen-html.sh` reads a `filters:` list from the `.md` frontmatter and resolves each bare name to `maint/filters/[name].lua` before invoking pandoc.

### source files (do not hand-edit the HTML they generate)
| source | generates | template | filters |
|---|---|---|---|
| `public/my-second-personal-site.me/index.md` | `index.html` | `index.template.html` (frontmatter) |
| `public/my-second-personal-site.me/cv/index.md` | `cv/index.html` | `cv/index.template.html` | none |
| `public/my-personal-site.me/index.md` | `index.html` | `index.template.html` | none (gen-toc updates article list section first) |
| `public/my-personal-site.me/projects/index.md` | `projects/index.html` | `projects/index.template.html` | none |
| `public/my-personal-site.me/writing/*/index.md` | `writing/*/index.html` | `writing/article.template.html` | `footnotes` (frontmatter) |

### URL convention
The worker serves `index.html` implicitly from directories — `public/site.me/foo/index.html` is served at `https://site.me/foo/` with no `.html` in the URL. All pages follow this pattern: a named directory containing `index.html`.

### adding a new generated page
Writing articles under `writing/` are picked up automatically — `common.mk` uses `find` recursively for `MD_SRCS`. Use `make new-article SLUG=my-slug` to scaffold a new article; no other Makefile changes needed.

For new **structural pages** (site sections, landing pages, etc. outside `writing/`):
1. Create `foo/index.md` with YAML frontmatter (`title` required)
2. Create `foo/index.template.html` alongside it (copy the closest existing template as a starting point)
3. If the page needs a lua filter, add a `filters:` list to the frontmatter — no Makefile changes needed
4. Add the new source file to the source files table above

For a **group of pages sharing one template** (like `writing/`): place the shared template in the group directory, then add a pattern rule to the site Makefile — the base `%.html: %.md %.template.html` rule won't fire for these because there's no sibling `.template.html` per page. Copy the `writing/%/index.html` rule as the reference:
```makefile
the/group/%/index.html: the/group/%/index.md the/group/shared.template.html
	MD=$(<) TEMPLATE=the/group/shared.template.html HTML=$(@) $(GEN_HTML_SH)
```
No other Makefile changes needed — `common.mk`'s recursive `find` picks up the `.md` files automatically.

Hand-authored HTML pages (no `.md` source) need only the directory and `index.html` — no Makefile or template changes required.

### pandoc invocation notes
- `--strip-comments` is passed on all invocations. HTML comments in `.md` source files are stripped from the output — use `<!-- /classname -->` labels freely on closing `:::` fences without polluting the rendered HTML.
- Fenced div closing convention: add `<!-- /classname -->` on the line after `:::` when the closing fence is more than two content lines from its opener. Applies to `.md` source files and Go text/templates alike.
- Lua filters are specified in frontmatter by bare name (e.g. `footnotes`); `gen-html.sh` resolves these to `maint/filters/[name].lua`. Filter changes are not tracked by Make — use `make -B` to force rebuild.

---

## author contracts

### frontmatter (all pandoc sources)
All .md sources use YAML frontmatter. `title` is the only required field.
* `title` — used in `<title>` and nav name-mark
* `subtitle` (optional) — used in nav name-mark subtitle span
* `date` (optional, articles) — used by site-tool gen-toc for article listing
* `description` (optional, articles) — used as RSS item description in feed.xml
* `filters` (optional) — list of lua filter bare names to apply; resolved to `maint/filters/[name].lua` by `gen-html.sh`
  ```yaml
  filters:
    - footnotes
    - foldout
  ```

### public/my-second-personal-site.me/index.md — fenced div conventions

* `## intro {.intro}` — produces `<section class="intro">` with h2 suppressed
* `## heading {#id}` — produces `<section id="id"><h2>heading</h2>`
* `::: lede` — single-paragraph div → `<p class="lede">`
* `::: story` — story block (collapsed by default)
* `::: {.story .open}` — story block open by default
* `::: ai-disclosure` — disclosure paragraph
* Inside a story block: first pure-italic paragraph → `.meta` div; first h3 → summary heading

### public/my-second-personal-site.me/cv/index.md — structural requirements
The CV uses structural CSS selectors (`.cv-body > h3 + p`, `.cv-body > h3 + p + p`, etc.) instead of classes.
**Blank lines are required between the three lines of each role entry:**
```markdown
### Company Name, City

*Company description.*

*Role title / date range*

- Bullet item
```
Without blank lines, pandoc merges adjacent lines into one `<p>`, breaking selector chaining.
Role titles use `*italic*` in source; CSS un-italics them via `h3 + p + p em { font-style: normal }`.

### dev articles — custom footnotes
The footnotes.lua filter handles bidirectional links for custom footnotes with arbitrary labels.

Convention: `[.LABEL]` — bracket + dot + label. LABEL is any non-empty string without `]`: symbols (`†`), numbers (`1`), words (`foo`), etc. Distinct from standard pandoc numbered footnotes (`[^n]`) and from link syntax. Grep-unique: `\[\.` finds all footnote references and definitions.

IDs are derived directly from the label: `[.foo]` → `id="fn-foo"` / `id="fnref-foo"`.

Author contract:
1. Place inline reference: `text[.foo]` (attached to preceding word — no space before bracket)
2. Place definition as a paragraph starting with `[.foo] Note text...` (typically under a `## Footnotes` section)
3. The filter adds all anchor links, backlinks, and ids automatically.

Articles may mix custom footnotes and standard pandoc numbered footnotes (`[^1]`, `[^2]`, etc.) in the same document — the filter only processes `[.LABEL]` markers and leaves `[^n]` footnotes to pandoc.

The filter also strips the `id` from the `## Footnotes` h2 to avoid conflicting with pandoc's own `<section id="footnotes">`.

---

## file inventory

### my-second-personal-site.me (public/my-second-personal-site.me/)
* style.css — canonical palette/aesthetic reference
* favicon.svg — "BH" monogram, slate-blue, glyphs as `<path>` (Helvetica Neue Bold via site-tool glyphs). SVG class `.accent` synced from `--accent` in style.css.
* favicon.ico — Safari fallback; regenerate with `make -C public/my-second-personal-site.me favicon`
* cv/style.css — CV-specific stylesheet (structural selectors, no classes)
* cv/MyNameCV.pdf — PDF download (manually maintained - upload markdown to google docs, download as pdf works)
* kreativ-architecture/index.html — pan/zoom SVG architecture viewer (hand-coded, bespoke JS)
* kreativ-architecture/KreativCrossPlatformEngineArchitecture.svg

### my-personal-site.me (public/my-personal-site.me/)
* style.css — canonical palette/aesthetic reference. `--accent-bright` defined here for favicon use (brighter than `--accent` for legibility on gray browser-tab chrome).
* favicon.svg — terminal "bh", amber/green accent, glyphs as `<path>` (Courier via site-tool glyphs). SVG classes `.bg` and `.accent-bright` synced from style.css.
* favicon.ico — Safari fallback; regenerate with `make -C public/my-personal-site.me favicon`
* rss.svg — RSS icon; uses `fill="currentColor"` as a CSS mask, no hardcoded colors, not a sync target.
* feed.xml — RSS feed; generated by site-tool gen-toc alongside index.html; served at /feed/
* index.md — article list section between markers is auto-updated by site-tool gen-toc; do not hand-edit that section
* writing/ — articles are self-describing via frontmatter; two exceptions:
  * ai-as-r2 — STUB (seed notes only, not ready)
  * ai-for-writing — unlisted (no date; linked from callout only, not in article index)
