package main

import (
	"flag"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

const (
	beginMarker = "<!-- ARTICLE-LIST-BEGIN -->"
	endMarker   = "<!-- ARTICLE-LIST-END -->"
)

func runGenTOC(args []string) {
	fs := flag.NewFlagSet("gen-toc", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: site-tool gen-toc [options]

Regenerates the article TOC in index.md and the RSS feed.xml from article
frontmatter. Articles without a date or with unlisted: true are excluded.

Options:
`)
		fs.PrintDefaults()
	}

	baseURL := fs.String("base-url", "", "Base URL of the site, e.g. https://example.com (required)")
	siteDesc := fs.String("site-desc", "", "Site description for RSS feed, e.g. \"Writing by Your Name\"")

	articlesDir := fs.String("articles-dir", "", "Path to writing articles directory (required)")
	indexFile := fs.String("index-file", "", "Path to index.md (required)")
	feedFile := fs.String("feed-file", "", "Path to RSS feed.xml (omit to skip RSS generation)")
	listTemplate := fs.String("list-template", "", "Path to article list markdown template (required)")
	siteTitle := fs.String("site-title", "", "Site title for RSS feed (defaults to bare base URL)")

	fs.Parse(args)

	if *baseURL == "" {
		fmt.Fprintln(os.Stderr, "error: --base-url is required")
		fs.Usage()
		os.Exit(1)
	}
	if *siteDesc == "" {
		fmt.Fprintln(os.Stderr, "error: --site-desc is required")
		fs.Usage()
		os.Exit(1)
	}
	if *articlesDir == "" {
		fmt.Fprintln(os.Stderr, "error: --articles-dir is required")
		fs.Usage()
		os.Exit(1)
	}
	if *indexFile == "" {
		fmt.Fprintln(os.Stderr, "error: --index-file is required")
		fs.Usage()
		os.Exit(1)
	}
	if *listTemplate == "" {
		fmt.Fprintln(os.Stderr, "error: --list-template is required")
		fs.Usage()
		os.Exit(1)
	}

	*baseURL = strings.TrimRight(*baseURL, "/")
	if *siteTitle == "" {
		*siteTitle = strings.TrimPrefix(strings.TrimPrefix(*baseURL, "https://"), "http://")
	}

	articles, err := collectArticles(*articlesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date > articles[j].Date
	})

	if err := updateIndexMD(*indexFile, articles, *listTemplate); err != nil {
		fmt.Fprintf(os.Stderr, "error updating index: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated %s with %d articles.\n", *indexFile, len(articles))

	if *feedFile == "" {
		fmt.Println("No --feed-file specified; skipping RSS feed generation.")
	} else {
		if err := writeFeedXML(*feedFile, articles, *baseURL, *siteTitle, *siteDesc); err != nil {
			fmt.Fprintf(os.Stderr, "error writing feed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %s with %d items.\n", *feedFile, len(articles))
	}
}

type article struct {
	Slug        string
	Date        string
	Title       string
	Subtitle    string
	Description string
	Tags        []string
}

func (a article) Display() string {
	if a.Subtitle != "" {
		return a.Title + ": " + a.Subtitle
	}
	return a.Title
}

func collectArticles(dir string) ([]article, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading articles dir %q: %w", dir, err)
	}

	var articles []article
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mdPath := filepath.Join(dir, e.Name(), "index.md")
		data, err := os.ReadFile(mdPath)
		if err != nil {
			continue
		}
		fm := parseFrontmatter(string(data))
		if fm["unlisted"] == "true" || fm["date"] == "" {
			continue
		}
		tags := fm["tags_list"] // populated by parseFrontmatter for list values
		var tagSlice []string
		if tags != "" {
			for _, t := range strings.Split(tags, "\x00") {
				if t != "" {
					tagSlice = append(tagSlice, t)
				}
			}
		}
		articles = append(articles, article{
			Slug:        e.Name(),
			Date:        fm["date"],
			Title:       fm["title"],
			Subtitle:    fm["subtitle"],
			Description: fm["description"],
			Tags:        tagSlice,
		})
	}
	return articles, nil
}

// parseFrontmatter parses simple YAML frontmatter without external dependencies.
// List values are stored with key "key_list" as NUL-separated strings.
func parseFrontmatter(text string) map[string]string {
	result := make(map[string]string)
	if !strings.HasPrefix(text, "---") {
		return result
	}
	end := strings.Index(text[3:], "\n---")
	if end == -1 {
		return result
	}
	fm := strings.TrimSpace(text[3 : 3+end])

	var currentListKey string
	for _, line := range strings.Split(fm, "\n") {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "- ") {
			if currentListKey != "" {
				result[currentListKey+"_list"] += strings.TrimPrefix(stripped, "- ") + "\x00"
			}
			continue
		}
		if idx := strings.Index(stripped, ":"); idx != -1 {
			key := strings.TrimSpace(stripped[:idx])
			value := strings.TrimSpace(stripped[idx+1:])
			value = strings.Trim(value, `"'`)
			if value != "" {
				result[key] = value
				currentListKey = ""
			} else {
				currentListKey = key
			}
		}
	}
	return result
}

func updateIndexMD(indexFile string, articles []article, tmplPath string) error {
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return fmt.Errorf("reading %q: %w", indexFile, err)
	}
	content := string(data)

	newList, err := buildArticleListMD(articles, tmplPath)
	if err != nil {
		return err
	}
	replacement := beginMarker + "\n" + newList + "\n" + endMarker

	pattern := regexp.MustCompile(`(?s)[ \t]*` + regexp.QuoteMeta(beginMarker) + `.*?` + regexp.QuoteMeta(endMarker))
	if !pattern.MatchString(content) {
		return fmt.Errorf("markers not found in %q", indexFile)
	}

	updated := pattern.ReplaceAllString(content, replacement)
	return os.WriteFile(indexFile, []byte(updated), 0644)
}

func buildArticleListMD(articles []article, tmplPath string) (string, error) {
	tmplData, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("reading list template %q: %w", tmplPath, err)
	}
	tmpl, err := template.New("article-list").Parse(string(tmplData))
	if err != nil {
		return "", fmt.Errorf("parsing list template %q: %w", tmplPath, err)
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, articles); err != nil {
		return "", fmt.Errorf("executing list template: %w", err)
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func toRFC2822(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	t = t.UTC()
	return t.Format("Mon, 02 Jan 2006 00:00:00 +0000")
}

func writeFeedXML(feedFile string, articles []article, baseURL, siteTitle, siteDesc string) error {
	feedURL := baseURL + "/feed/"
	now := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 +0000")

	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	b.WriteString("<rss version=\"2.0\" xmlns:atom=\"http://www.w3.org/2005/Atom\">\n")
	b.WriteString("  <channel>\n")
	b.WriteString("    <title>" + html.EscapeString(siteTitle) + "</title>\n")
	b.WriteString("    <link>" + baseURL + "/</link>\n")
	b.WriteString("    <description>" + html.EscapeString(siteDesc) + "</description>\n")
	b.WriteString("    <language>en-us</language>\n")
	b.WriteString("    <lastBuildDate>" + now + "</lastBuildDate>\n")
	b.WriteString("    <atom:link href=\"" + feedURL + "\" rel=\"self\" type=\"application/rss+xml\"/>\n")

	for _, a := range articles {
		url := baseURL + "/writing/" + a.Slug + "/"
		display := html.EscapeString(a.Display())
		b.WriteString("    <item>\n")
		b.WriteString("      <title>" + display + "</title>\n")
		b.WriteString("      <link>" + url + "</link>\n")
		b.WriteString("      <guid isPermaLink=\"true\">" + url + "</guid>\n")
		b.WriteString("      <pubDate>" + toRFC2822(a.Date) + "</pubDate>\n")
		if a.Description != "" {
			b.WriteString("      <description>" + html.EscapeString(a.Description) + "</description>\n")
		}
		b.WriteString("    </item>\n")
	}

	b.WriteString("  </channel>\n")
	b.WriteString("</rss>\n")

	newContent := b.String()

	// Skip write if only lastBuildDate changed — avoids spurious git diffs on every build.
	if existing, err := os.ReadFile(feedFile); err == nil {
		if stripLastBuildDate(string(existing)) == stripLastBuildDate(newContent) {
			return nil
		}
	}

	return os.WriteFile(feedFile, []byte(newContent), 0644)
}

var lastBuildDateRE = regexp.MustCompile(`(?m)^\s*<lastBuildDate>.*</lastBuildDate>\n?`)

func stripLastBuildDate(xml string) string {
	return lastBuildDateRE.ReplaceAllString(xml, "")
}
