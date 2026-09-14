# site-tool — Agent Guide

Go CLI binary providing this repo's site maintenance utilities: `gen-toc`,
`sync-svg-colors`, `glyphs`, `version`. Read [SPEC.md](SPEC.md) for the exact
CLI contract (flags, exit codes, guarantees) and [ARCHITECTURE.md](ARCHITECTURE.md)
for module layout, invariants, and the procedure for adding a subcommand —
this file does not restate either.

Read `/Users/agent-user/.claude/CLAUDE.md` (or your project's equivalent
global instructions) and `../../agents/go-coder.md` before making changes
here — this file assumes both are already in context and covers only what
they don't: e.g. never invoke `go build`/`go test` directly when a Make
target below does the same thing, and dispatch build/test work to a coder
agent per the global coding policy.

## Module

```
module site-tool   (go 1.23.0)
```

Go 1.23 is a floor, not a pin — building with a newer toolchain is fine.

## Build, test, dist

Run from `maint/site-tool/` directly, or from `maint/` as
`make site-tool` / `make site-tool-test` / `make site-tool-dist` / etc., or
from the repo root as `make site-tool` / `make site-tool-test` / ... (root
`Makefile` forwards to `maint/`).

| target | output | notes |
|---|---|---|
| `build` | `../bin/site-tool` (or `.exe`) | current host platform only; for development |
| `test` | — | `go test ./...`; `VERBOSE=1` for `-v` |
| `dist` | `../bin/site-tool-{os}-{arch}[.exe]`, 6 targets | depends on `test` — never ships from a red tree |
| `clean` | — | removes the local dev build only |
| `nuke` | — | removes all of `../bin/` |

**`../bin/` (i.e. `maint/bin/`) is entirely gitignored in this repository** —
both the local dev build and every dist binary. This repo always builds
site-tool from source; there is no committed binary here to fall back on.
(The public sibling template repo removes that `.gitignore` line and commits
the dist binaries via Git LFS instead — ARCHITECTURE.md §4 has the full
rationale. Don't assume that behavior applies here.)

## How the parent project invokes this

Every call site goes through `maint/site-tool.sh` (platform-detecting
wrapper — ARCHITECTURE.md §4), never a direct binary path:

- `common.mk` defines `SITE_TOOL_SH` for all site Makefiles.
- `favicon.mk`'s `sync-svg-colors` target calls
  `$(SITE_TOOL_SH) sync-svg-colors --css style.css favicon.svg`.
- Each site's `Makefile` (e.g. `public/bennherrera.dev/Makefile`) calls
  `$(SITE_TOOL_SH) gen-toc --base-url ... --articles-dir writing ...` to
  regenerate `index.md`'s article list and `feed.xml`.

If you change a subcommand's flags, grep the repo for `site-tool.sh` call
sites and update them in the same change — SPEC.md is the contract, but
these Makefile lines are what actually exercises it.
