package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseCSSColorMap(t *testing.T) {
	css := `:root {
  --bg: #112233;
  --accent: #ff7f00;
}
/* a commented-out override that must NOT win:
   --bg: #000000;
*/
@media (prefers-color-scheme: light) {
  :root { --accent: #00ff00; }
}`
	got := parseCSSColorMap(css)

	if got[""]["bg"] != "#112233" {
		t.Errorf("root bg = %q, want #112233", got[""]["bg"])
	}
	if got[""]["accent"] != "#ff7f00" {
		t.Errorf("root accent = %q, want #ff7f00", got[""]["accent"])
	}
	light := "@media (prefers-color-scheme: light)"
	if got[light]["accent"] != "#00ff00" {
		t.Errorf("light accent = %q, want #00ff00", got[light]["accent"])
	}
}

func TestExtractMediaBlocks(t *testing.T) {
	css := `.x { fill: #111 }
@media (prefers-color-scheme: dark) { .a { fill: #222 } .b { fill: #333 } }
.y { fill: #444 }
@media (min-width: 600px) { .c { fill: #555 } }`
	blocks := extractMediaBlocks(css)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 media blocks, got %d", len(blocks))
	}
	if blocks[0].query != "@media (prefers-color-scheme: dark)" {
		t.Errorf("block 0 query = %q", blocks[0].query)
	}
	// brace-depth tracking must capture both inner rules
	if !strings.Contains(blocks[0].inner, ".a") || !strings.Contains(blocks[0].inner, ".b") {
		t.Errorf("block 0 inner missing rules: %q", blocks[0].inner)
	}
	if blocks[1].query != "@media (min-width: 600px)" {
		t.Errorf("block 1 query = %q", blocks[1].query)
	}
}

func TestDefaultContent(t *testing.T) {
	css := "A@media (x) { inner }B"
	blocks := extractMediaBlocks(css)
	got := defaultContent(css, blocks)
	if got != "AB" {
		t.Errorf("defaultContent = %q, want %q", got, "AB")
	}
}

func TestEffectivePalette(t *testing.T) {
	colors := svgColorMap{
		"":                                     {"bg": "#000000", "accent": "#ff0000"},
		"@media (prefers-color-scheme: light)": {"accent": "#00ff00"},
	}
	got := effectivePalette(colors, "@media (prefers-color-scheme: light)")
	want := map[string]string{"bg": "#000000", "accent": "#00ff00"} // root overlaid with light
	if !reflect.DeepEqual(got, want) {
		t.Errorf("effectivePalette = %v, want %v", got, want)
	}

	// unknown query falls back to root defaults only
	gotRoot := effectivePalette(colors, "@media (no-such)")
	wantRoot := map[string]string{"bg": "#000000", "accent": "#ff0000"}
	if !reflect.DeepEqual(gotRoot, wantRoot) {
		t.Errorf("effectivePalette(unknown) = %v, want %v", gotRoot, wantRoot)
	}
}

func TestTopLevelColorQuery(t *testing.T) {
	dark := mediaBlock{query: "@media (prefers-color-scheme: dark)"}
	light := mediaBlock{query: "@media (prefers-color-scheme: light)"}
	other := mediaBlock{query: "@media (min-width: 600px)"}

	cases := []struct {
		name   string
		blocks []mediaBlock
		want   string
	}{
		{"only dark override → top-level is light", []mediaBlock{dark}, "@media (prefers-color-scheme: light)"},
		{"only light override → top-level is dark", []mediaBlock{light}, "@media (prefers-color-scheme: dark)"},
		{"both present → dark", []mediaBlock{dark, light}, "@media (prefers-color-scheme: dark)"},
		{"neither → dark", []mediaBlock{other}, "@media (prefers-color-scheme: dark)"},
		{"empty → dark", nil, "@media (prefers-color-scheme: dark)"},
	}
	for _, c := range cases {
		if got := topLevelColorQuery(c.blocks); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestReplaceClassColors(t *testing.T) {
	colors := map[string]string{"accent": "#ff7f00"}

	t.Run("replaces matching class", func(t *testing.T) {
		got, changed := replaceClassColors(".accent { fill: #000000 }", colors)
		if !changed || got != ".accent { fill: #ff7f00 }" {
			t.Errorf("got %q changed=%v", got, changed)
		}
	})
	t.Run("leaves unknown class untouched", func(t *testing.T) {
		got, changed := replaceClassColors(".other { fill: #000000 }", colors)
		if changed || got != ".other { fill: #000000 }" {
			t.Errorf("got %q changed=%v", got, changed)
		}
	})
	t.Run("no change when color already matches", func(t *testing.T) {
		_, changed := replaceClassColors(".accent { fill: #ff7f00 }", colors)
		if changed {
			t.Errorf("expected changed=false when color already matches")
		}
	})
}

func TestUpdateElementAttrs(t *testing.T) {
	rules := map[string]map[string]string{
		"bg": {"fill": "#112233"},
	}
	in := `<rect class="bg" fill="#000000"/><path class="accent" fill="#000000"/>`
	got, changed := updateElementAttrs(in, rules)
	if !changed {
		t.Fatal("expected changed=true")
	}
	if !strings.Contains(got, `<rect class="bg" fill="#112233"/>`) {
		t.Errorf("bg fill not updated: %q", got)
	}
	if !strings.Contains(got, `<path class="accent" fill="#000000"/>`) {
		t.Errorf("non-matching class should be untouched: %q", got)
	}
}

// End-to-end: a small SVG with a light-mode override block synced against a
// dark-rooted CSS palette. Exercises updateStyleContent + element attr sync.
func TestSyncSVGColors(t *testing.T) {
	colors := svgColorMap{
		"":                                     {"bg": "#112233", "accent": "#ff7f00"},
		"@media (prefers-color-scheme: light)": {"accent": "#00ff00"},
	}
	svg := `<svg><style>
.bg { fill: #000000 }
.accent { fill: #000000 }
@media (prefers-color-scheme: light) {
  .accent { fill: #000000 }
}
</style>
<rect class="bg" fill="#000000"/>
<path class="accent" fill="#000000"/>
</svg>`

	out, changed, err := syncSVGColors(svg, colors)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	// Top-level rules represent the CSS :root (dark) palette.
	if !strings.Contains(out, ".bg { fill: #112233 }") {
		t.Errorf("top-level .bg not synced to root:\n%s", out)
	}
	if !strings.Contains(out, ".accent { fill: #ff7f00 }") {
		t.Errorf("top-level .accent not synced to root:\n%s", out)
	}
	// The @media (light) block uses the light override.
	if !strings.Contains(out, "#00ff00") {
		t.Errorf("light-mode override not synced:\n%s", out)
	}
	// Element attributes follow the top-level (default) rules.
	if !strings.Contains(out, `<rect class="bg" fill="#112233"/>`) {
		t.Errorf("rect attr not synced:\n%s", out)
	}
	if !strings.Contains(out, `<path class="accent" fill="#ff7f00"/>`) {
		t.Errorf("path attr not synced:\n%s", out)
	}

	// Idempotence: a second pass makes no further changes.
	out2, changed2, err := syncSVGColors(out, colors)
	if err != nil {
		t.Fatal(err)
	}
	if changed2 {
		t.Errorf("second sync should be a no-op")
	}
	if out2 != out {
		t.Errorf("second sync altered output")
	}
}
