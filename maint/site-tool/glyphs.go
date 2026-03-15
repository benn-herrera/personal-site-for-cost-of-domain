package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func runGlyphs(args []string) {
	fs := flag.NewFlagSet("glyphs", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: site-tool glyphs --font <path> [--chars <chars>] [options]
       site-tool glyphs --font <path> --list

Extract font glyphs as SVG <path> elements. Outputs one <path> per character,
positioned using the same coordinate system as SVG text (x = left edge,
baseline = y). Drop-in replacements for <text> elements.

Options:
`)
		fs.PrintDefaults()
	}

	fontPath := fs.String("font", "", "Font file path (.ttf, .otf, .ttc) (required)")
	chars := fs.String("chars", "", "Characters to extract (required unless --list)")
	listFaces := fs.Bool("list", false, "List all faces in a .ttc file with their index and name")
	fontSize := fs.Float64("font-size", 26, "Equivalent CSS font-size in SVG user units")
	xPos := fs.Float64("x", 0, "X position for the first character")
	baseline := fs.Float64("baseline", 23, "Y position of the text baseline in SVG coords")
	fontNumber := fs.Int("font-number", -1, "Font index within a .ttc collection (overrides -face)")
	faceName := fs.String("face", "", "Case-insensitive pattern to match against face name (exact match preferred)")

	fs.Parse(args)

	if *fontPath == "" {
		fmt.Fprintln(os.Stderr, "error: --font is required")
		fs.Usage()
		os.Exit(1)
	}

	data, err := os.ReadFile(*fontPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot read font file: %v\n", err)
		os.Exit(1)
	}

	// Build face list — try collection first, fall back to single font.
	type faceEntry struct {
		index int
		font  *sfnt.Font
	}
	var faces []faceEntry

	if col, err := sfnt.ParseCollection(data); err == nil {
		n := col.NumFonts()
		for i := 0; i < n; i++ {
			f, err := col.Font(i)
			if err != nil {
				continue
			}
			faces = append(faces, faceEntry{i, f})
		}
	} else {
		f, err := sfnt.Parse(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot parse font: %v\n", err)
			os.Exit(1)
		}
		faces = []faceEntry{{0, f}}
	}

	strippedName := func(f *sfnt.Font) string {
		full, _ := f.Name(nil, sfnt.NameIDFull)
		family, _ := f.Name(nil, sfnt.NameIDFamily)
		s := strings.TrimPrefix(full, family)
		s = strings.TrimSpace(s)
		if s == "" {
			return full
		}
		return s
	}

	if *listFaces {
		for _, fe := range faces {
			fmt.Printf("  %2d  %s\n", fe.index, strippedName(fe.font))
		}
		return
	}

	if *chars == "" {
		fmt.Fprintln(os.Stderr, "error: --chars is required unless --list is specified")
		fs.Usage()
		os.Exit(1)
	}

	// Resolve which face to use.
	var selectedFont *sfnt.Font
	selectedName := ""

	if *fontNumber >= 0 {
		for _, fe := range faces {
			if fe.index == *fontNumber {
				selectedFont = fe.font
				selectedName = strippedName(fe.font)
				break
			}
		}
		if selectedFont == nil {
			fmt.Fprintf(os.Stderr, "error: no face with index %d\n", *fontNumber)
			os.Exit(1)
		}
	} else if *faceName != "" {
		pattern := strings.ToLower(*faceName)
		var exact, partial []faceEntry
		for _, fe := range faces {
			name := strings.ToLower(strippedName(fe.font))
			if name == pattern {
				exact = append(exact, fe)
			} else if strings.Contains(name, pattern) {
				partial = append(partial, fe)
			}
		}
		matches := exact
		if len(matches) == 0 {
			matches = partial
		}
		switch len(matches) {
		case 0:
			fmt.Fprintf(os.Stderr, "error: no face matching %q\navailable faces:\n", *faceName)
			for _, fe := range faces {
				fmt.Fprintf(os.Stderr, "  %2d  %s\n", fe.index, strippedName(fe.font))
			}
			os.Exit(1)
		case 1:
			selectedFont = matches[0].font
			selectedName = strippedName(matches[0].font)
		default:
			fmt.Fprintf(os.Stderr, "error: %d faces match %q — be more specific:\n", len(matches), *faceName)
			for _, fe := range matches {
				fmt.Fprintf(os.Stderr, "  %2d  %s\n", fe.index, strippedName(fe.font))
			}
			os.Exit(1)
		}
	} else {
		selectedFont = faces[0].font
		selectedName = strippedName(faces[0].font)
	}

	// Extract glyphs.
	upm := int32(selectedFont.UnitsPerEm())
	ppem := fixed.Int26_6(upm << 6) // unitsPerEm in 26.6 fixed point
	scale := *fontSize / float64(upm)
	x := *xPos

	var buf sfnt.Buffer
	for _, r := range *chars {
		gi, err := selectedFont.GlyphIndex(&buf, r)
		if err != nil || gi == 0 {
			fmt.Printf("  <!-- warning: %q U+%04X not found in font -->\n", r, r)
			continue
		}

		segs, err := selectedFont.LoadGlyph(&buf, gi, ppem, nil)
		if err != nil {
			fmt.Printf("  <!-- warning: could not load glyph for %q: %v -->\n", r, err)
			continue
		}

		advance, err := selectedFont.GlyphAdvance(&buf, gi, ppem, 0)
		if err != nil {
			fmt.Printf("  <!-- warning: could not get advance for %q -->\n", r)
			continue
		}

		pathData := segmentsToSVGPath(segs, scale, x, *baseline)
		advanceSVG := float64(advance) / 64.0 * scale

		fmt.Printf("  <!-- %q x=%.2f baseline=%g advance=%.2f font-size=%g face=%q -->\n",
			r, x, *baseline, advanceSVG, *fontSize, selectedName)
		fmt.Printf("  <path d=%q />\n", pathData)

		x += advanceSVG
	}
}

func fmtCoord(v float64) string {
	// Format like %g but avoid scientific notation for the coordinate ranges
	// we expect in SVG favicons (roughly 0–100).
	if v == math.Trunc(v) {
		return fmt.Sprintf("%g", v)
	}
	return fmt.Sprintf("%g", v)
}

func transformPt(rawX, rawY fixed.Int26_6, scale, xOffset, baseline float64) (float64, float64) {
	fx := float64(rawX) / 64.0
	fy := float64(rawY) / 64.0
	return xOffset + fx*scale, baseline - fy*scale
}

func segmentsToSVGPath(segs []sfnt.Segment, scale, xOffset, baseline float64) string {
	var b strings.Builder
	first := true

	for _, seg := range segs {
		switch seg.Op {
		case sfnt.SegmentOpMoveTo:
			if !first {
				b.WriteString("Z")
			}
			first = false
			px, py := transformPt(seg.Args[0].X, seg.Args[0].Y, scale, xOffset, baseline)
			fmt.Fprintf(&b, "M%s %s", fmtCoord(px), fmtCoord(py))

		case sfnt.SegmentOpLineTo:
			px, py := transformPt(seg.Args[0].X, seg.Args[0].Y, scale, xOffset, baseline)
			fmt.Fprintf(&b, "L%s %s", fmtCoord(px), fmtCoord(py))

		case sfnt.SegmentOpQuadTo:
			x1, y1 := transformPt(seg.Args[0].X, seg.Args[0].Y, scale, xOffset, baseline)
			px, py := transformPt(seg.Args[1].X, seg.Args[1].Y, scale, xOffset, baseline)
			fmt.Fprintf(&b, "Q%s %s %s %s", fmtCoord(x1), fmtCoord(y1), fmtCoord(px), fmtCoord(py))

		case sfnt.SegmentOpCubeTo:
			x1, y1 := transformPt(seg.Args[0].X, seg.Args[0].Y, scale, xOffset, baseline)
			x2, y2 := transformPt(seg.Args[1].X, seg.Args[1].Y, scale, xOffset, baseline)
			px, py := transformPt(seg.Args[2].X, seg.Args[2].Y, scale, xOffset, baseline)
			fmt.Fprintf(&b, "C%s %s %s %s %s %s",
				fmtCoord(x1), fmtCoord(y1),
				fmtCoord(x2), fmtCoord(y2),
				fmtCoord(px), fmtCoord(py))
		}
	}

	if !first {
		b.WriteString("Z")
	}
	return b.String()
}
