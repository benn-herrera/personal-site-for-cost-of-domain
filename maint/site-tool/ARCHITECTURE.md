# site-tool ARCHITECTURE

Design reference and single source of truth for site-tool's structure,
invariants, and extension procedure. [SPEC.md](SPEC.md) is the externally
observable CLI contract — the *what a caller may rely on* — and does not
restate the material below.

---

## 1. Overview

site-tool is a single Go binary bundling the personal-site build's
maintenance utilities as subcommands, replacing what used to be two
standalone Python scripts. It has no runtime beyond the OS: no interpreter
to install, no `pip` environment, one static executable per platform.

It is invoked almost exclusively through `maint/site-tool.sh` (§4), never
built or run ad hoc — see `CONVENTIONS.md` for the build/test/dist recipes.

## 2. Module layout

```
main.go              entry point: version constant, usage constant, command dispatch
gentoc.go             gen-toc subcommand
syncsvgcolors.go       sync-svg-colors subcommand
glyphs.go              glyphs subcommand
*_test.go              table-driven unit tests, one file per subcommand
Makefile               build/test/dist/clean/nuke targets
```

**One-subcommand-per-file.** Every subcommand lives in its own `<name>.go`,
all `package main` — there is no internal package split, because there is
no code shared between subcommands worth abstracting (see §3, dependency
policy: even the frontmatter parser and the CSS color parser, though both
"parsers," share no logic and stay local to `gentoc.go` and
`syncsvgcolors.go` respectively).

Each subcommand is entered through exactly one function,
`run<Name>(args []string)`, called from the `switch` in `main.go`. That
function owns everything: flag definition, validation, and execution. There
is no shared subcommand framework or registry — `main.go`'s switch statement
is the whole of it, deliberately, because three subcommands do not justify
one.

## 3. Invariants

Violating any of these is a blocking defect.

- **Stdlib-first; one external dependency.** `golang.org/x/image` (its
  `font/sfnt` and `math/fixed` packages) is the sole direct external
  dependency, needed for `glyphs` — Go's standard library has no font/glyph
  parser. `golang.org/x/text` appears in `go.mod` only as `x/image`'s own
  transitive dependency; nothing in this module imports it directly. **No
  YAML parser** — `gen-toc`'s frontmatter reader (SPEC.md §6) is
  hand-written against the narrow subset article frontmatter actually uses,
  because a real YAML dependency buys generality this tool has no use for.
  **No XML library** — `gen-toc`'s RSS writer (SPEC.md §2) is string
  concatenation with `html.EscapeString` at every text boundary, because
  the output shape is fixed and small enough that a document builder would
  be more code, not less.
- **Each subcommand owns its own `flag.FlagSet`** (not the global
  `flag.CommandLine`), with a custom `Usage` func. This is what lets
  `main.go`'s dispatch be a plain switch without flag collisions between
  subcommands.
- **Required-flag validation happens after `fs.Parse`, by hand** — Go's
  `flag` package has no required-flag concept. A missing required flag
  prints `error: --<flag> is required` (or equivalent) followed by
  `fs.Usage()`, then `os.Exit(1)`. See SPEC.md §1 for the full error/exit
  contract.
- **Errors to stderr, output to stdout**, no exceptions.
- **Idempotence where the operation implies it.** `sync-svg-colors`
  reports `no changes` and performs no write when a file already matches
  the resolved palette (SPEC.md §3); `gen-toc`'s feed writer skips the
  write entirely when the only delta is `<lastBuildDate>` (SPEC.md §2).
- **Deterministic output.** The parent project's `make verify` byte-compares
  committed generated content against a fresh rebuild (see `CONVENTIONS.md`'s
  pointer to that pipeline) — any nondeterminism in `gen-toc` or
  `sync-svg-colors` output breaks that gate. Map iteration order is never
  allowed to leak into emitted output; `parseCSSColorMap`/`svgColorMap`
  values are read back out by explicit key, and article/palette ordering is
  sourced from sorted slices, never ranged-over maps.

## 4. Design decisions

**Single multi-subcommand Go binary, replacing two Python scripts.**
site-tool supersedes `glyph-to-path.py` and `gen-dev-index.py` (both
deleted). The prior scripts required a Python interpreter plus `fonttools`
for glyph extraction — a real dependency footprint for a project whose only
other toolchain requirement was `pandoc` and (optionally) `wrangler`. Folding
both scripts' logic into one Go binary with `golang.org/x/image` for font
parsing eliminates the Python/fonttools dependency entirely: a contributor
building this project from scratch needs Go (to build site-tool from
source) or nothing at all (to run a pre-built binary), never Python.
One binary with subcommands, rather than three small binaries, because
three near-trivial `main` packages sharing no code would multiply build/dist
bookkeeping (§6) for no isolation benefit — nothing about `gen-toc` needs to
ship, version, or fail independently of `glyphs`.

**Why `site-tool.sh` exists.** The parent Makefiles (`common.mk`,
`favicon.mk`, per-site `Makefile`s) need one stable path to invoke, on any
of the supported host platforms, without each call site encoding
`$(GOOS)-$(GOARCH)` logic itself. `maint/site-tool.sh` centralizes that:
it detects host OS/arch, tries the local dev build
(`maint/bin/site-tool[.exe]`, from `make build`) first, falls back to the
platform-named dist binary (`maint/bin/site-tool-<os>-<arch>[.exe]`), and
validates whichever it finds actually runs (`site-tool version`) before
handing off — catching a stale or platform-incompatible binary at the call
site instead of deep inside a Make recipe's failure output. Every Makefile
that shells out to site-tool goes through this one script
(`common.mk`'s `SITE_TOOL_SH`).

**Distribution split — this repo builds from source; the sibling repo ships
binaries.** In *this* repository, `maint/bin/` is entirely `.gitignore`d
(dev build and every per-platform dist name alike) — nothing under it is
committed here, and `make site-tool` (from source) is how this repo obtains
a usable binary. The public sibling template repository
(`personal-site-for-cost-of-domain`) removes that one `.gitignore` line and
commits the six cross-compiled dist binaries via Git LFS instead, so a
non-Go user cloning the *template* can start editing sample content
immediately. Go source ships in both repos either way. This split exists
because the two repos have different audiences: this repo's contributors
are expected to have (or install) Go; the template's target user is
explicitly "no toolchain required."

**Cross-compile matrix.** `windows/amd64`, `windows/arm64`, `linux/amd64`,
`linux/arm64`, `darwin/amd64`, `darwin/arm64` — six targets, built by
`make dist` (`Makefile`'s `dist` target, one `GOOS=... GOARCH=... go build`
line per target). Pure Go with no cgo dependency in either `glyphs.go` (the
one package with a non-stdlib import) makes every target buildable from any
one host — no per-platform build machine needed. `dist` depends on `test`
(`Makefile` line `dist: test`), so a red test suite never produces distro
binaries.

## 5. Adding a new subcommand

1. Create `<name>.go` in this directory, `package main`.
2. Implement `run<Name>(args []string)`:
   - own `flag.FlagSet` with a custom `Usage` func (§3)
   - validate required flags after `fs.Parse`; on failure, print
     `error: ...` + `fs.Usage()` to stderr, `os.Exit(1)` (§3, SPEC.md §1)
   - errors to stderr, output to stdout (§3)
3. Add `case "<name>": run<Name>(os.Args[2:])` to the switch in `main.go`.
4. Add the command to the `usage` constant in `main.go`.
5. Bump `version` in `main.go`.
6. Add `<name>_test.go` with table-driven tests covering the new logic
   (existing `*_test.go` files are the pattern to match). `make test` must
   stay green.
7. Add the subcommand's contract to `SPEC.md` (flags, exit behavior,
   guarantees) and, if it introduces a new design decision or dependency,
   to this file's §3/§4.
8. Run `make dist` to rebuild all six platform binaries (`maint/bin/`,
   locally gitignored either way — see §4's distribution split for where
   these ship from).
