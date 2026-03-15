# site-tool — Agent Guide

Go CLI binary providing site maintenance utilities. Invoked via `maint/site-tool.sh`, which detects the host platform and dispatches to the appropriate pre-built binary in `maint/bin/`.

## module

```
module site-tool   (go 1.23.0)
```

One external dependency: `golang.org/x/image` (for `sfnt` font parsing in `glyphs.go`). All other subcommands use only the standard library — no YAML parser, no XML library.

## file layout

| file | purpose |
|---|---|
| `main.go` | entry point; `version` constant; command dispatch switch |
| `gentoc.go` | `gen-toc` subcommand |
| `syncsvgcolors.go` | `sync-svg-colors` subcommand |
| `glyphs.go` | `glyphs` subcommand |
| `Makefile` | build/dist/clean targets |

All files are `package main`. Each subcommand lives in its own file with a single `run<Name>(args []string)` entry function.

## Go toolchain

Building from source requires the Go toolchain. Pre-built binaries for all supported platforms are checked in via Git LFS, so most users will never need to build — the `make build` / `make dist` targets are for development and when adding new subcommands.

If the user does need to build and does not have Go installed, help them install it first:
* macOS: `brew install go`
* Linux: `sudo apt install golang-go` or `sudo pacman -S go`
* Windows: `winget install -e --id GoLang.Go`
* Or directly from https://go.dev/dl/

Verify with `go version` — requires Go 1.23+.

## build targets

| target | output | notes |
|---|---|---|
| `build` | `maint/bin/site-tool` (or `.exe`) | current platform only; gitignored; for development |
| `dist` | `maint/bin/site-tool-{os}-{arch}[.exe]` | all 6 targets; checked in via Git LFS |
| `clean` | — | removes local build only |
| `nuke` | — | removes entire `maint/bin/` |

Cross-compile matrix: `windows/amd64`, `windows/arm64`, `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`.

Run from `maint/` as `make site-tool` / `make site-tool-dist`, or directly from this directory.

## CLI conventions

- Each subcommand uses its own `flag.FlagSet` with a custom `Usage` func.
- Required flags are validated after `fs.Parse`; missing required flags print an error + usage and `os.Exit(1)`.
- Errors go to stderr; normal output goes to stdout.
- All subcommands are idempotent where applicable (e.g. `sync-svg-colors` reports "no changes" if already up to date).

---

## subcommands

### gen-toc

**File:** `gentoc.go`

Scans an articles directory, reads YAML frontmatter from each `writing/*/index.md`, and:
1. Updates the article list section in `index.md` between `<!-- ARTICLE-LIST-BEGIN -->` and `<!-- ARTICLE-LIST-END -->` markers using a Go `text/template`.
2. Optionally writes an RSS 2.0 `feed.xml`.

**Flags:** `--base-url` (required), `--site-desc` (required), `--articles-dir` (required), `--index-file` (required), `--list-template` (required), `--feed-file` (optional), `--site-title` (defaults to bare base URL).

**Article inclusion:** articles without a `date` frontmatter field, or with `unlisted: true`, are silently excluded.

**Article sort:** reverse chronological by `date` string (lexicographic — `YYYY-MM-DD` format sorts correctly).

**`article` struct fields available in the list template:**

| field | source |
|---|---|
| `.Slug` | directory name |
| `.Date` | frontmatter `date` |
| `.Title` | frontmatter `title` |
| `.Subtitle` | frontmatter `subtitle` |
| `.Description` | frontmatter `description` |
| `.Tags` | frontmatter `tags` list |
| `.Display()` | `Title: Subtitle` if subtitle present, else `Title` |

**Frontmatter parser:** custom, no external deps. Scalar values: `key: value`. List values: a key followed by `- item` lines; stored internally as NUL-separated strings under `key_list`. The parser strips leading/trailing quotes from scalar values.

**RSS:** RSS 2.0 with `xmlns:atom` self-link. Dates converted from `YYYY-MM-DD` to RFC 2822 (`Mon, 02 Jan 2006 00:00:00 +0000`). Item `<description>` only emitted if `description` frontmatter is present. `<lastBuildDate>` reflects the current time but the file is only written when article content actually changed — if the new output differs from the existing file only in `<lastBuildDate>`, the file is left untouched to avoid spurious git diffs on every build.

---

### sync-svg-colors

**File:** `syncsvgcolors.go`

Reads CSS custom property hex color values from a stylesheet and syncs them into SVG `<style>` blocks and element `fill`/`stroke` attributes.

**Flags:** `--css <style.css>` (required), then one or more SVG file paths as positional arguments.

**CSS → SVG mapping:** SVG class name = CSS custom property name without `--`. E.g. CSS `--accent: #ff7f00` syncs to SVG `.accent { fill: #ff7f00 }`. Only hex colors are matched (`#[0-9a-fA-F]{3,8}`).

**`@media` handling:** the parser extracts `@media` blocks using brace-depth tracking (handles nesting). The effective palette for a given context is the CSS `:root` defaults merged with any context-specific overrides.

**Top-level SVG style context detection:** determined by which `@media (prefers-color-scheme:...)` blocks exist in the SVG's own `<style>`:
- SVG has only a `dark` override block → top-level rules represent light mode → syncs light CSS palette to them.
- Otherwise → top-level rules represent the CSS `:root` default mode → syncs that palette.

**Element attribute sync:** after updating the `<style>` block, also syncs `fill`/`stroke` attributes on element tags whose class matches a top-level style rule (keeps inline attributes in sync with the style block).

---

### glyphs

**File:** `glyphs.go`

Extracts font glyphs as SVG `<path>` elements from `.ttf`, `.otf`, or `.ttc` font files. Output goes to stdout — copy the `<path>` elements into an SVG manually.

**Flags:** `--font` (required), `--chars` (required unless `--list`), `--list` (list all faces in a .ttc), `--font-size` (default 26), `--x` (default 0), `--baseline` (default 23), `--font-number` (index), `--face` (name pattern).

**Face selection priority:** `--font-number` > `--face` (exact match, then partial) > first face.

**Coordinate system:** x = left edge of first character, baseline = y. Matches SVG `<text>` positioning — paths are drop-in replacements.

**Output:** one `<path d="..."/>` per character, preceded by an HTML comment with metrics (x position, baseline, advance, font-size, face name). Advance accumulates left-to-right across characters.

**Dependency:** `golang.org/x/image/font/sfnt` and `golang.org/x/image/math/fixed`.

---

## adding a new subcommand

1. Create `<name>.go` in this directory, `package main`.
2. Implement `run<Name>(args []string)` following the existing pattern:
   - Own `flag.FlagSet` with custom `Usage`
   - Validate required flags after `fs.Parse`; exit 1 with error + usage on failure
   - Errors to stderr, output to stdout
3. Add a `case "<name>":` in the switch in `main.go` calling `run<Name>(os.Args[2:])`.
4. Add the command to the `usage` constant in `main.go`.
5. Bump `version` in `main.go`.
6. Run `make dist` to rebuild all platform binaries and update `maint/bin/` (LFS).
7. Update the parent `AGENTS.md` pipeline components section and targets table if relevant.
