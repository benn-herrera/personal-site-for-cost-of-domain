package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	t.Run("scalars with quote stripping", func(t *testing.T) {
		fm := parseFrontmatter(`---
title: "Hello World"
date: 2026-01-02
subtitle: 'A Sub'
---
body text`)
		want := map[string]string{
			"title":    "Hello World",
			"date":     "2026-01-02",
			"subtitle": "A Sub",
		}
		for k, v := range want {
			if fm[k] != v {
				t.Errorf("%q = %q, want %q", k, fm[k], v)
			}
		}
	})

	t.Run("value containing a colon", func(t *testing.T) {
		// split is on the first colon only
		fm := parseFrontmatter(`---
title: "Ratio 16:9 explained"
---`)
		if fm["title"] != "Ratio 16:9 explained" {
			t.Errorf("title = %q, want %q", fm["title"], "Ratio 16:9 explained")
		}
	})

	t.Run("list values stored NUL-separated under key_list", func(t *testing.T) {
		fm := parseFrontmatter(`---
title: T
tags:
  - alpha
  - beta
  - gamma
---`)
		if fm["title"] != "T" {
			t.Errorf("title = %q, want T", fm["title"])
		}
		if got, want := fm["tags_list"], "alpha\x00beta\x00gamma\x00"; got != want {
			t.Errorf("tags_list = %q, want %q", got, want)
		}
		// the list key itself should not be set as a scalar
		if _, ok := fm["tags"]; ok {
			t.Errorf("tags should not be set as a scalar value")
		}
	})

	t.Run("scalar after a list resets list context", func(t *testing.T) {
		fm := parseFrontmatter(`---
tags:
  - a
title: After
---`)
		if fm["tags_list"] != "a\x00" {
			t.Errorf("tags_list = %q, want %q", fm["tags_list"], "a\x00")
		}
		if fm["title"] != "After" {
			t.Errorf("title = %q, want After", fm["title"])
		}
	})

	t.Run("unlisted flag", func(t *testing.T) {
		fm := parseFrontmatter(`---
title: T
unlisted: true
---`)
		if fm["unlisted"] != "true" {
			t.Errorf("unlisted = %q, want true", fm["unlisted"])
		}
	})

	t.Run("no frontmatter returns empty map", func(t *testing.T) {
		fm := parseFrontmatter("# Just a heading\n\nbody")
		if len(fm) != 0 {
			t.Errorf("expected empty map, got %v", fm)
		}
	})

	t.Run("unterminated frontmatter returns empty map", func(t *testing.T) {
		fm := parseFrontmatter("---\ntitle: T\nno closing fence")
		if len(fm) != 0 {
			t.Errorf("expected empty map, got %v", fm)
		}
	})
}

func TestArticleDisplay(t *testing.T) {
	cases := []struct {
		name string
		a    article
		want string
	}{
		{"with subtitle", article{Title: "Title", Subtitle: "Sub"}, "Title: Sub"},
		{"without subtitle", article{Title: "Title"}, "Title"},
	}
	for _, c := range cases {
		if got := c.a.Display(); got != c.want {
			t.Errorf("%s: Display() = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestToRFC2822(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"2026-01-02", "Fri, 02 Jan 2026 00:00:00 +0000"},
		{"2026-02-28", "Sat, 28 Feb 2026 00:00:00 +0000"},
		{"not-a-date", "not-a-date"}, // invalid passes through unchanged
	}
	for _, c := range cases {
		if got := toRFC2822(c.in); got != c.want {
			t.Errorf("toRFC2822(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestStripLastBuildDate(t *testing.T) {
	a := "<rss>\n    <lastBuildDate>Mon, 01 Jan 2026 00:00:00 +0000</lastBuildDate>\n  <item/>\n</rss>\n"
	b := "<rss>\n    <lastBuildDate>Tue, 02 Feb 2027 11:22:33 +0000</lastBuildDate>\n  <item/>\n</rss>\n"
	if stripLastBuildDate(a) != stripLastBuildDate(b) {
		t.Errorf("feeds differing only in lastBuildDate should strip equal:\n%q\n%q",
			stripLastBuildDate(a), stripLastBuildDate(b))
	}
	if strings.Contains(stripLastBuildDate(a), "lastBuildDate") {
		t.Errorf("lastBuildDate line not stripped: %q", stripLastBuildDate(a))
	}
}

func TestBuildArticleListMD(t *testing.T) {
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "list.md")
	if err := os.WriteFile(tmpl, []byte("{{range .}}- {{.Slug}}: {{.Display}}\n{{end}}"), 0644); err != nil {
		t.Fatal(err)
	}
	articles := []article{
		{Slug: "a", Title: "Alpha", Subtitle: "first"},
		{Slug: "b", Title: "Beta"},
	}
	got, err := buildArticleListMD(articles, tmpl)
	if err != nil {
		t.Fatal(err)
	}
	want := "- a: Alpha: first\n- b: Beta" // trailing newline trimmed
	if got != want {
		t.Errorf("buildArticleListMD = %q, want %q", got, want)
	}
}

func TestUpdateIndexMD(t *testing.T) {
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "list.md")
	if err := os.WriteFile(tmpl, []byte("{{range .}}- {{.Slug}}\n{{end}}"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("replaces between markers, preserves surrounding content", func(t *testing.T) {
		index := filepath.Join(dir, "index.md")
		content := "intro line\n" + beginMarker + "\nstale\n" + endMarker + "\noutro line\n"
		if err := os.WriteFile(index, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		err := updateIndexMD(index, []article{{Slug: "x"}, {Slug: "y"}}, tmpl)
		if err != nil {
			t.Fatal(err)
		}
		out, _ := os.ReadFile(index)
		s := string(out)
		if !strings.Contains(s, "- x\n- y") {
			t.Errorf("new list not present:\n%s", s)
		}
		if strings.Contains(s, "stale") {
			t.Errorf("stale content not replaced:\n%s", s)
		}
		if !strings.HasPrefix(s, "intro line\n") || !strings.HasSuffix(s, "outro line\n") {
			t.Errorf("surrounding content not preserved:\n%s", s)
		}
	})

	t.Run("errors when markers absent", func(t *testing.T) {
		index := filepath.Join(dir, "nomarkers.md")
		if err := os.WriteFile(index, []byte("no markers here"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := updateIndexMD(index, nil, tmpl); err == nil {
			t.Errorf("expected error when markers missing")
		}
	})
}

func TestWriteFeedXML(t *testing.T) {
	dir := t.TempDir()
	feed := filepath.Join(dir, "feed.xml")
	articles := []article{
		{Slug: "x", Date: "2026-01-02", Title: "Title <X>", Description: "desc & stuff"},
	}

	if err := writeFeedXML(feed, articles, "https://example.com", "Example", "Site Desc"); err != nil {
		t.Fatal(err)
	}
	s := readFile(t, feed)
	for _, want := range []string{
		"<item>",
		"https://example.com/writing/x/",
		"Title &lt;X&gt;",  // title HTML-escaped
		"desc &amp; stuff", // description HTML-escaped
		"Fri, 02 Jan 2026 00:00:00 +0000",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("feed missing %q:\n%s", want, s)
		}
	}

	t.Run("skips rewrite when only lastBuildDate would differ", func(t *testing.T) {
		// Plant a sentinel lastBuildDate; an unchanged-content rebuild must leave it.
		planted := strings.Replace(s, "<lastBuildDate>", "<lastBuildDate>SENTINEL", 1)
		// give it a recognizable body too
		planted = lastBuildDateRE.ReplaceAllString(planted, "    <lastBuildDate>SENTINEL</lastBuildDate>\n")
		if err := os.WriteFile(feed, []byte(planted), 0644); err != nil {
			t.Fatal(err)
		}
		if err := writeFeedXML(feed, articles, "https://example.com", "Example", "Site Desc"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(readFile(t, feed), "SENTINEL") {
			t.Errorf("expected rebuild to be skipped (SENTINEL preserved), but file was rewritten")
		}
	})

	t.Run("rewrites when article content changes", func(t *testing.T) {
		changed := append(articles, article{Slug: "y", Date: "2026-03-04", Title: "Second"})
		if err := writeFeedXML(feed, changed, "https://example.com", "Example", "Site Desc"); err != nil {
			t.Fatal(err)
		}
		out := readFile(t, feed)
		if strings.Contains(out, "SENTINEL") {
			t.Errorf("expected rewrite to drop SENTINEL, got:\n%s", out)
		}
		if !strings.Contains(out, "https://example.com/writing/y/") {
			t.Errorf("new item missing after content change:\n%s", out)
		}
	})
}

func TestCollectArticles(t *testing.T) {
	dir := t.TempDir()
	writeArticle := func(slug, body string) {
		d := filepath.Join(dir, slug)
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "index.md"), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeArticle("good", `---
title: Good One
date: 2026-02-03
subtitle: Sub
description: Desc
tags:
  - x
  - y
---
body`)
	writeArticle("nodate", "---\ntitle: No Date\n---\nbody")
	writeArticle("hidden", "---\ntitle: Hidden\ndate: 2026-01-01\nunlisted: true\n---\nbody")
	// a stray file at the top level should be ignored
	if err := os.WriteFile(filepath.Join(dir, "loose.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := collectArticles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 listed article, got %d: %+v", len(got), got)
	}
	a := got[0]
	if a.Slug != "good" || a.Title != "Good One" || a.Date != "2026-02-03" ||
		a.Subtitle != "Sub" || a.Description != "Desc" {
		t.Errorf("unexpected article fields: %+v", a)
	}
	if len(a.Tags) != 2 || a.Tags[0] != "x" || a.Tags[1] != "y" {
		t.Errorf("tags = %v, want [x y]", a.Tags)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
