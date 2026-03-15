package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

const usage = `site-tool — site maintenance utilities

Usage:
    site-tool <command> [options]

Commands:
    glyphs            Extract font glyphs as SVG <path> elements
    gen-toc           Regenerate article TOC HTML and RSS feed
    sync-svg-colors   Sync CSS custom property colors into SVG files
    version           Print version and exit

Run 'site-tool <command> -help' for command-specific options.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "glyphs":
		runGlyphs(os.Args[2:])
	case "gen-toc":
		runGenTOC(os.Args[2:])
	case "sync-svg-colors":
		runSyncSVGColors(os.Args[2:])
	case "version":
		fmt.Println("site-tool " + version)
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}
