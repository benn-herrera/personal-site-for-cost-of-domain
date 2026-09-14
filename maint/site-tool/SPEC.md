# site-tool SPEC

The contract: what site-tool guarantees a caller (the parent build, a hand
invocation, or a clean-room reimplementation) may rely on. [ARCHITECTURE.md](ARCHITECTURE.md)
is the design reference — the *how* and *why*, and the single source for
file layout and extension procedure this document does not restate.

---

## 1. Process contract

```
site-tool <command> [options]
```

No global flags. Each subcommand owns its own flag set (ARCHITECTURE.md §2).

**Exit codes.** `0` on success. `1` on: no command given, an unknown
command, a missing required flag, or any runtime failure (unreadable file,
unwritable output, malformed input the subcommand cannot process). No other
exit code is used.

**Output channels.** Errors go to stderr, normally prefixed `error:` or
`error <doing-what>:` (the exact wording varies by call site — callers may
match on exit code, not on message text). Normal output goes to stdout.
`site-tool` with no arguments, or an unrecognized command, prints usage to
stderr and exits `1`. `-h` / `--help` / `help` prints usage to stdout and
exits `0`.

## 2. `site-tool gen-toc`

```
site-tool gen-toc --base-url URL --site-desc DESC --articles-dir DIR
                   --index-file FILE --list-template FILE
                   [--feed-file FILE] [--site-title TITLE]
```

Scans an articles directory for frontmatter, rewrites the article-list
section of a markdown index file in place, and optionally writes an RSS feed.

**Flags:**

| Flag | Required | Default | Notes |
|---|---|---|---|
| `--base-url` | yes | — | trailing `/` stripped before use |
| `--site-desc` | yes | — | RSS channel `<description>` |
| `--articles-dir` | yes | — | directory of `<slug>/index.md` article sources |
| `--index-file` | yes | — | markdown file rewritten in place |
| `--list-template` | yes | — | Go `text/template` source, pandoc markdown output |
| `--feed-file` | no | unset | omit to skip RSS generation entirely |
| `--site-title` | no | bare `--base-url` (scheme stripped) | RSS channel `<title>` |

A missing required flag prints `error: --<flag> is required` plus usage to
stderr and exits `1`, checked in the order listed above (first missing flag
reported, not all at once).

**Article inclusion.** A directory in `--articles-dir` is a candidate only if
it holds `index.md` (unreadable or absent → silently skipped, not an error).
Within that file's frontmatter, an article is **excluded** — silently, no
diagnostic — when either holds: no `date` key, or `unlisted: true` (string
comparison, exact).

**Sort.** Included articles sort by `date` string, descending, using plain
Go string comparison (`>`). `YYYY-MM-DD` sorts correctly under this scheme;
any other date format does not.

**Article fields available to `--list-template`:**

| Field | Source |
|---|---|
| `.Slug` | the article's directory name |
| `.Date` | frontmatter `date` |
| `.Title` | frontmatter `title` |
| `.Subtitle` | frontmatter `subtitle` |
| `.Description` | frontmatter `description` |
| `.Tags` | frontmatter `tags` list |
| `.Display` | method: `"Title: Subtitle"` if `Subtitle` is non-empty, else `"Title"` |

The template is executed with the full article slice (`{{range .}}`) as its
only input — no site-level fields (base URL, site title) are exposed to it.

**Index-file rewrite contract.** `--index-file` must already contain, once,
the literal markers:

```
<!-- ARTICLE-LIST-BEGIN -->
<!-- ARTICLE-LIST-END -->
```

Everything from the first `BEGIN` marker (including any leading run of
spaces/tabs on its line) through the first `END` marker that follows it is
replaced with `BEGIN\n<rendered template, trailing newlines trimmed>\nEND`.
Content before and after the marker span is preserved byte-for-byte. If the
marker pair is not found, the file is **not written** and the command exits
`1`. The rewritten file is written with mode `0644`.

**RSS contract** (only when `--feed-file` is set). RSS 2.0, one `<channel>`:

- `<title>` = `--site-title` (or its derived default), HTML-escaped.
- `<link>` = `--base-url/`.
- `<description>` = `--site-desc`, HTML-escaped.
- `<language>` = `en-us`, fixed.
- `<atom:link href="<base-url>/feed/" rel="self" type="application/rss+xml"/>`
  — the feed's own self-reference, per the `xmlns:atom` namespace declared on
  `<rss>`.
- One `<item>` per included article, in the same sorted order as the index:
  `<title>` = `.Display`, HTML-escaped; `<link>`/`<guid isPermaLink="true">`
  = `<base-url>/writing/<slug>/`; `<pubDate>` = frontmatter `date` converted
  to RFC 2822 (`Mon, 02 Jan 2006 00:00:00 +0000`, always UTC/`+0000`; a
  `date` that does not parse as `YYYY-MM-DD` passes through **unconverted**,
  producing an RSS feed with a non-conformant `<pubDate>` rather than
  failing the build); `<description>` emitted, HTML-escaped, **only when**
  frontmatter `description` is present.
- `<lastBuildDate>` = current UTC time at generation.

**`<lastBuildDate>`-only-diff suppression.** Before writing, the newly
rendered feed and the existing on-disk feed (if any) are each compared with
their own `<lastBuildDate>` line stripped out. If the stripped forms are
equal, **the file is not written at all** — the existing file, old
`<lastBuildDate>` included, is left exactly as it was. This is what keeps a
content-unchanged rebuild from producing a git diff on every run. A feed file
that does not yet exist, or whose stripped content differs, is written in
full (mode `0644`).

## 3. `site-tool sync-svg-colors`

```
site-tool sync-svg-colors --css FILE SVG [SVG ...]
```

Syncs hex color values from CSS custom properties into one or more SVG
files' `<style>` blocks and element `fill`/`stroke` attributes, in place.

**Flags:** `--css FILE` (required). At least one positional SVG path is
also required — zero positional args is an error, not a no-op. Missing
`--css` or zero SVG args each print `error: ...` plus usage to stderr and
exit `1`. A read failure on `--css` or any SVG file, or a write failure on
any SVG file, prints `error ...: <cause>` to stderr and exits `1`
immediately (files after the failing one in argument order are not
processed).

**Per-file result line** (stdout): `updated:    <path>` or
`no changes: <path>`.

**CSS → SVG class mapping.** An SVG class name matches a CSS custom
property with the same name minus the `--` prefix: CSS `--accent: #ff7f00`
syncs `.accent { fill: #ff7f00 }` / `.accent { stroke: #ff7f00 }` rules and
`class="accent"` element attributes. Only hex colors are recognized —
pattern `#[0-9a-fA-F]{3,8}` (a bare digit-count check, not a validator: a
5- or 7-digit run matches the regex even though it is not a legal CSS hex
color; malformed CSS input is not rejected, just possibly mis-synced).
`/* ... */` CSS comments (including multi-line) are stripped before
parsing, so a commented-out property definition never wins.

**`@media` handling.** `@media` blocks are located by brace-depth tracking
(nesting-safe) and keyed by their full query string, e.g.
`@media (prefers-color-scheme: light)`. The **effective palette** for a
given query is the CSS `:root`-level (top-level, non-`@media`) property set
merged with that query's own overrides — the query's values win on key
collision, and a query with no CSS block of its own falls back to the
`:root` set alone.

**Top-level SVG context detection** (`topLevelColorQuery`). The SVG's own
`<style>` block is scanned for `@media (prefers-color-scheme: ...)` blocks:

- Only a `dark` override present → the SVG's top-level (non-`@media`) rules
  are read as the **light**-mode palette, and synced against the CSS light
  query's effective palette.
- Anything else (a `light` override present, both present, neither
  present, or no `@media` blocks at all) → top-level rules are synced
  against the CSS **dark** query's effective palette.

**Known limitation.** The "anything else" branch always resolves to dark —
there is no read of the CSS file's own `:root` values to determine which
mode they actually represent. This is correct only because both sites in
this repo are dark-rooted (`:root` holds the dark palette, with a `light`
`@media` override). A CSS palette that is light-rooted (`:root` = light
values, `dark` as the override) would have its favicon light/dark sync
inverted by this heuristic. Verified against `topLevelColorQuery` in
`syncsvgcolors.go` as of this writing — it is a hard-coded fallback, not
context-sensitive.

**Style block sync.** Each top-level and `@media`-scoped region of the
`<style>` block has its `.class { fill|stroke: #hex }` rules rewritten to
the resolved palette's value for that class, case-insensitive comparison
(no write occurs if already equal). A class with no entry in the resolved
palette is left untouched.

**Element attribute sync.** After the style block is (possibly) rewritten,
every element start tag carrying a `class="..."` attribute is scanned; for
each space-separated class present in the resulting **top-level** style
rules, that element's existing `fill=` and/or `stroke=` attribute value is
rewritten to match. An element whose tag has **no** existing `fill`/`stroke`
attribute is not modified — attributes are synced, never added.

**Idempotence.** A second run over already-synced output makes no further
changes (`updated` → `no changes`) and produces byte-identical file content.

## 4. `site-tool glyphs`

```
site-tool glyphs --font FILE [--chars STR] [options]
site-tool glyphs --font FILE --list
```

Extracts glyph outlines from a font file as SVG `<path>` elements, one per
character in `--chars`, to stdout. Not applied to any file — output is
meant to be copied by hand into an SVG.

**Flags:**

| Flag | Required | Default | Notes |
|---|---|---|---|
| `--font` | yes | — | `.ttf`, `.otf`, or `.ttc` |
| `--chars` | yes unless `--list` | — | characters to extract, in iteration order |
| `--list` | no | `false` | list every face in the font/collection and exit; `--chars` not required |
| `--font-size` | no | `26` | CSS-equivalent font-size, SVG user units |
| `--x` | no | `0` | left edge of the first character |
| `--baseline` | no | `23` | Y position of the text baseline, SVG coords |
| `--font-number` | no | `-1` | face index within a `.ttc`; `-1` = unset |
| `--face` | no | `""` | case-insensitive name pattern |

`--font` missing → error + usage, exit `1`. Without `--list`, `--chars`
missing → error + usage, exit `1`. An unreadable or unparseable `--font`
file → error, exit `1`.

**Face selection priority**, first match wins:

1. `--font-number N` set (`N >= 0`): the face at that exact index. No face
   at that index → error naming the index, exit `1`.
2. `--face PATTERN` set: faces whose (family-prefix-stripped) name equals
   the pattern case-insensitively; if none, faces whose name *contains* the
   pattern case-insensitively. Exactly one match → selected. Zero matches →
   error listing every available face (index + name), exit `1`. More than
   one match → error listing the matching faces, exit `1`.
3. Neither set: the first face in file order (index 0 for a plain
   `.ttf`/`.otf`; the first collection member for a `.ttc`).

A plain (non-collection) font is treated as a one-face list; `--font-number`
and `--face` apply to it the same way.

**Coordinate system.** `--x` is the left edge of the first character;
`--baseline` is that text's baseline Y — the same convention SVG `<text>`
uses, so emitted `<path>` elements are drop-in replacements for a `<text>`
element at the same position. Units are scaled from font units by
`--font-size / unitsPerEm`.

**Output**, per character in `--chars`, in order: an HTML comment with
metrics, then the path:

```
  <!-- 'X' x=12.34 baseline=23 advance=15.6 font-size=26 face="Bold" -->
  <path d="M... Z" />
```

A character not present in the selected face, or whose glyph fails to load
or measure, emits a `<!-- warning: ... -->` comment in its place and no
`<path>` — the run continues (not a fatal error) and `--x` does **not**
advance for that character. `advance` accumulates left-to-right: each
character's `x` is the running total of prior characters' advances plus the
initial `--x`. Path coordinates are formatted with the shortest
round-tripping decimal representation, never scientific notation, regardless
of magnitude.

**Dependency:** `golang.org/x/image/font/sfnt` (parsing) and
`golang.org/x/image/math/fixed` (26.6 fixed-point coordinates) — see
ARCHITECTURE.md §2 for why this is the one sanctioned external dependency.

## 5. `site-tool version`

`site-tool version` prints `site-tool <version>` to stdout and exits `0`.
No flags.

## 6. Frontmatter parser contract

Used only by `gen-toc` (§2), to read each article's `index.md`. Custom,
no external YAML dependency (ARCHITECTURE.md §2) — a deliberately narrow
subset of YAML frontmatter, not a general parser.

- A file must open with `---` as its literal first three bytes and contain
  a following line consisting of `---` (i.e. `\n---`), or the whole file
  parses as an **empty** map — this is not treated as an error at any call
  site.
- Within the block, each line is trimmed of leading/trailing whitespace.
- A line beginning `- ` is a list item, appended (with a trailing `\x00`)
  to `<key>_list` where `<key>` is the most recent scalar-line key whose
  value was empty (see below). A list item encountered with no such key
  pending is silently dropped.
- Any other line containing `:` is a scalar: split at the **first** `:`
  only (so a value itself containing `:` is preserved whole, e.g.
  `title: "Ratio 16:9 explained"`). The value is trimmed, then has any
  leading and trailing characters in the set `"'` stripped (independently
  per end — mismatched or doubled quote characters are stripped too, this
  is not a matched-pair grammar).
    - Non-empty value after stripping → stored under `<key>` as a scalar,
      and any pending list key is cleared.
    - Empty value → no scalar is stored; `<key>` becomes the pending list
      key for subsequent `- ` lines.
- A line with no `:` and not a `- ` item is ignored.

**What this parser does NOT support** (a real contract boundary for
article authors): multi-line/block scalars (`|`, `>`); inline flow
collections (`tags: [a, b]`); nested maps or nested lists; type coercion
(`true`/numbers stay strings — `unlisted: true` is compared as the literal
string `"true"`); comments (a `#` inside a scalar line is not stripped and
becomes part of the value); anchors, aliases, or multi-document `---`
separators beyond the first closing one; duplicate-key merge semantics
(a repeated key simply overwrites the prior scalar).

---

## Needs ruling

- Error-message prefix is inconsistent across call sites (`error: X is
  required` vs. `error reading X: ...` vs. unprefixed `unknown command:
  ...`). Callers should match on exit code, not message text, until this is
  normalized.
- `sync-svg-colors`'s hex-color regex (§3) accepts any 3–8 digit run, not
  just CSS's legal lengths (3/4/6/8) — whether a 5- or 7-digit match should
  be rejected rather than silently synced is unruled.
