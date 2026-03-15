package main

import (
	"flag"
	"fmt"
	"maps"
	"os"
	"regexp"
	"strings"
)

func runSyncSVGColors(args []string) {
	fs := flag.NewFlagSet("sync-svg-colors", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: site-tool sync-svg-colors --css <style.css> <svg-file> [<svg-file> ...]

Syncs CSS custom property color values into SVG <style> blocks and element
fill/stroke attributes. SVG class names must match CSS custom property names
(without the -- prefix). @media contexts are matched by query string.

Options:
`)
		fs.PrintDefaults()
	}

	cssFile := fs.String("css", "", "CSS file containing custom property color values (required)")
	fs.Parse(args)

	if *cssFile == "" {
		fmt.Fprintln(os.Stderr, "error: --css is required")
		fs.Usage()
		os.Exit(1)
	}

	svgFiles := fs.Args()
	if len(svgFiles) == 0 {
		fmt.Fprintln(os.Stderr, "error: at least one SVG file is required")
		fs.Usage()
		os.Exit(1)
	}

	cssData, err := os.ReadFile(*cssFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", *cssFile, err)
		os.Exit(1)
	}
	colors := parseCSSColorMap(string(cssData))

	for _, svgFile := range svgFiles {
		svgData, err := os.ReadFile(svgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", svgFile, err)
			os.Exit(1)
		}
		updated, changed, err := syncSVGColors(string(svgData), colors)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error processing %s: %v\n", svgFile, err)
			os.Exit(1)
		}
		if changed {
			if err := os.WriteFile(svgFile, []byte(updated), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "error writing %s: %v\n", svgFile, err)
				os.Exit(1)
			}
			fmt.Printf("updated:    %s\n", svgFile)
		} else {
			fmt.Printf("no changes: %s\n", svgFile)
		}
	}
}

// svgColorMap maps media query string → CSS property name (without --) → hex color.
// The "" key holds default (:root) values.
type svgColorMap map[string]map[string]string

var cssPropRe = regexp.MustCompile(`--([a-zA-Z][a-zA-Z0-9-]*)\s*:\s*(#[0-9a-fA-F]{3,8})\b`)
var cssCommentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)

// parseCSSColorMap extracts CSS custom property hex values from a stylesheet,
// grouped by media query context.
func parseCSSColorMap(cssText string) svgColorMap {
	cssText = cssCommentRe.ReplaceAllString(cssText, "")
	result := svgColorMap{}
	blocks := extractMediaBlocks(cssText)

	for _, m := range cssPropRe.FindAllStringSubmatch(defaultContent(cssText, blocks), -1) {
		if result[""] == nil {
			result[""] = make(map[string]string)
		}
		result[""][m[1]] = m[2]
	}
	for _, b := range blocks {
		for _, m := range cssPropRe.FindAllStringSubmatch(b.inner, -1) {
			if result[b.query] == nil {
				result[b.query] = make(map[string]string)
			}
			result[b.query][m[1]] = m[2]
		}
	}
	return result
}

// mediaBlock holds a parsed @media block with position info for reconstruction.
type mediaBlock struct {
	start      int    // position of '@' in the source
	end        int    // position after closing '}'
	query      string // e.g., "@media (prefers-color-scheme: light)"
	inner      string // content inside the braces
	innerStart int    // position of first char inside '{'
	innerEnd   int    // position of closing '}'
}

// extractMediaBlocks extracts all @media blocks from CSS text using brace-depth tracking.
func extractMediaBlocks(text string) []mediaBlock {
	var blocks []mediaBlock
	i := 0
	for {
		idx := strings.Index(text[i:], "@media")
		if idx == -1 {
			break
		}
		start := i + idx
		i = start + len("@media")

		braceIdx := strings.Index(text[i:], "{")
		if braceIdx == -1 {
			break
		}
		query := "@media " + strings.TrimSpace(text[i:i+braceIdx])
		i += braceIdx + 1
		innerStart := i

		depth := 1
		for i < len(text) && depth > 0 {
			switch text[i] {
			case '{':
				depth++
			case '}':
				depth--
			}
			i++
		}
		innerEnd := i - 1 // points at the closing '}'

		blocks = append(blocks, mediaBlock{
			start:      start,
			end:        i,
			query:      query,
			inner:      text[innerStart:innerEnd],
			innerStart: innerStart,
			innerEnd:   innerEnd,
		})
	}
	return blocks
}

// defaultContent returns text with all @media blocks removed.
func defaultContent(text string, blocks []mediaBlock) string {
	var sb strings.Builder
	prev := 0
	for _, b := range blocks {
		sb.WriteString(text[prev:b.start])
		prev = b.end
	}
	sb.WriteString(text[prev:])
	return sb.String()
}

var (
	svgStyleBlockRe = regexp.MustCompile(`(?s)(<style[^>]*>)(.*?)(</style>)`)
	svgClassColorRe = regexp.MustCompile(`(\.([a-zA-Z][a-zA-Z0-9-]*)\s*\{\s*(?:fill|stroke)\s*:\s*)(#[0-9a-fA-F]{3,8})`)
	svgElementTagRe = regexp.MustCompile(`(?s)<[a-zA-Z][a-zA-Z0-9]*\b[^>]*/?>`)
	svgClassAttrRe  = regexp.MustCompile(`\bclass="([^"]*)"`)
)

// syncSVGColors updates an SVG's <style> block and element fill/stroke attributes
// to match the provided CSS color map.
func syncSVGColors(svgText string, colors svgColorMap) (string, bool, error) {
	changed := false

	styleIdx := svgStyleBlockRe.FindStringSubmatchIndex(svgText)
	if styleIdx == nil {
		return svgText, false, nil
	}
	styleContent := svgText[styleIdx[4]:styleIdx[5]]

	updatedStyle, styleChanged := updateStyleContent(styleContent, colors)
	if styleChanged {
		changed = true
		svgText = svgText[:styleIdx[4]] + updatedStyle + svgText[styleIdx[5]:]
		styleContent = updatedStyle
	}

	// Sync element attributes to the current default style values.
	defRules := defaultStyleRules(styleContent)
	if updatedSVG, attrChanged := updateElementAttrs(svgText, defRules); attrChanged {
		changed = true
		svgText = updatedSVG
	}

	return svgText, changed, nil
}

// effectivePalette returns the effective color palette for a given @media query context,
// starting with the CSS :root defaults and overlaying the query-specific overrides.
// If the query has no CSS definitions, returns just the root defaults.
func effectivePalette(colors svgColorMap, query string) map[string]string {
	palette := make(map[string]string, len(colors[""]))
	maps.Copy(palette, colors[""])
	maps.Copy(palette, colors[query])
	return palette
}

// topLevelColorQuery returns the CSS @media query that represents the color mode used
// by the SVG's top-level (non-@media) style rules. If the SVG's style block contains
// only a @media (prefers-color-scheme: dark) override — meaning the top-level rules
// represent the light-mode baseline — this returns the light query. Otherwise the
// top-level rules represent the CSS :root default (dark) mode and this returns the dark
// query (which falls back to root if no explicit dark block exists in the CSS).
func topLevelColorQuery(blocks []mediaBlock) string {
	hasDark, hasLight := false, false
	for _, b := range blocks {
		if strings.Contains(b.query, "prefers-color-scheme: dark") {
			hasDark = true
		}
		if strings.Contains(b.query, "prefers-color-scheme: light") {
			hasLight = true
		}
	}
	if hasDark && !hasLight {
		return "@media (prefers-color-scheme: light)"
	}
	return "@media (prefers-color-scheme: dark)"
}

// updateStyleContent applies CSS colors to a style block, handling @media contexts.
// Top-level rules use the effective palette for the mode they represent (detected from
// the SVG's own @media blocks); @media rules use the effective palette for their query
// (CSS explicit overrides merged with CSS :root fallback).
func updateStyleContent(styleText string, colors svgColorMap) (string, bool) {
	changed := false
	blocks := extractMediaBlocks(styleText)
	topColors := effectivePalette(colors, topLevelColorQuery(blocks))

	var sb strings.Builder
	prev := 0
	for _, b := range blocks {
		// Top-level content before this @media block.
		seg := styleText[prev:b.start]
		if updated, c := replaceClassColors(seg, topColors); c {
			changed = true
			seg = updated
		}
		sb.WriteString(seg)

		// @media header — unchanged.
		sb.WriteString(styleText[b.start:b.innerStart])

		// @media inner content: effective palette for this query (explicit overrides + root fallback).
		inner := b.inner
		if updated, c := replaceClassColors(inner, effectivePalette(colors, b.query)); c {
			changed = true
			inner = updated
		}
		sb.WriteString(inner)

		// Closing '}'.
		sb.WriteString(styleText[b.innerEnd:b.end])
		prev = b.end
	}

	// Remaining top-level content after last @media block.
	seg := styleText[prev:]
	if updated, c := replaceClassColors(seg, topColors); c {
		changed = true
		seg = updated
	}
	sb.WriteString(seg)

	if changed {
		return sb.String(), true
	}
	return styleText, false
}

// replaceClassColors replaces hex color values in CSS class rules when the class
// name matches a key in cssColors.
func replaceClassColors(text string, cssColors map[string]string) (string, bool) {
	changed := false
	result := svgClassColorRe.ReplaceAllStringFunc(text, func(match string) string {
		m := svgClassColorRe.FindStringSubmatch(match)
		prefix, className, currentColor := m[1], m[2], m[3]
		newColor, ok := cssColors[className]
		if !ok || strings.EqualFold(currentColor, newColor) {
			return match
		}
		changed = true
		return prefix + newColor
	})
	return result, changed
}

// defaultStyleRules extracts class → { "fill": color, "stroke": color } from
// the top-level (non-@media) rules of a style block.
func defaultStyleRules(styleText string) map[string]map[string]string {
	blocks := extractMediaBlocks(styleText)
	rules := make(map[string]map[string]string)
	for _, m := range svgClassColorRe.FindAllStringSubmatch(defaultContent(styleText, blocks), -1) {
		prefix, className, color := m[1], m[2], m[3]
		prop := "fill"
		if strings.Contains(prefix, "stroke") {
			prop = "stroke"
		}
		if rules[className] == nil {
			rules[className] = make(map[string]string)
		}
		rules[className][prop] = color
	}
	return rules
}

// updateElementAttrs updates fill/stroke attributes on SVG element tags whose
// class matches a rule in the provided map.
func updateElementAttrs(svgText string, rules map[string]map[string]string) (string, bool) {
	if len(rules) == 0 {
		return svgText, false
	}
	changed := false
	result := svgElementTagRe.ReplaceAllStringFunc(svgText, func(tag string) string {
		cm := svgClassAttrRe.FindStringSubmatch(tag)
		if cm == nil {
			return tag
		}
		updatedTag := tag
		for _, cls := range strings.Fields(cm[1]) {
			propColors, ok := rules[cls]
			if !ok {
				continue
			}
			for prop, newColor := range propColors {
				attrRe := regexp.MustCompile(`(` + prop + `=")#[0-9a-fA-F]{3,8}(")`)
				if newTag := attrRe.ReplaceAllString(updatedTag, "${1}"+newColor+`${2}`); newTag != updatedTag {
					changed = true
					updatedTag = newTag
				}
			}
		}
		return updatedTag
	})
	return result, changed
}
