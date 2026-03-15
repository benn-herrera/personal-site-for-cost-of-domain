# site-tool
A tool for maintenance operations like maintaining article table of contents and RSS feed that require custom behavior better implemented in a real coding language instead of shell script

## Why Go
The two major contenders for this tool were python and go. Because the tool needed to be delivered to users conveniently that pretty muched stabbed python right in the virtual environment. Go lets you cross compile for all targets from any dev host. A single utility binary for 6 possibler configurations. For people who want to modify the tool Go is a single dependency that is easy to install on all platforms. Again, no virtual environment and interpreter versioning mess.

The tool binary is not referenced directly from makefiles. It is wrapped by `site-tool.sh` which checks for a local build first, then checks for the appropriate os/arch version.

## Make targets
* `build`: 
  * builds `../bin/site-tool` (or `site-tool.exe` on Windows)
  * git ignored
  * this executable takes precedence over the prebuilt os/arch named binaries in site-tool.sh
* `clean`: removes `../bin/site-tool` (or `site-tool.exe` on Windows)
* `dist`: builds the os/arch named binaries for distribution (version controlled via git LFS)
* `nuke`: removes everything in `../bin/`

## Tool commnds
All all commands support `--help` for usage

* `gen-toc`
  * primary command, used for updatding table of contents lists
  * optionally generates RSS feed XML
* `glyphs`
  * utility for converting characters from local fonts to SVG paths
  * `<text>` elements in SVG may not render consistently across machines due to local font availability
  * allows for previewing SVG designs with text and then converting to paths for reliability
* `sync-svg-colors`
  * utility for keeping favicon.svg light/dark mode colors in sync with style.css
  * matches 'class' attributes in SVG to color variables in style.css
* `version`
  * astonishingly enough, prints the tool version and exits
  * used in `site-tool.sh` to verify os/arch compatibility with current system

## Details
See AGENTS.md
