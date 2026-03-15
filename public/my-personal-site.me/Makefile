MAINT_DIR := ../../maint
# must come first
include $(MAINT_DIR)/common.mk

FEED_XML              := feed.xml
ARTICLE_LIST_TEMPLATE := templates/gen-toc-article-list.md
ARTICLE_TEMPLATE      := writing/article.template.html

.PHONY: public new-article sync-domain-names

# HTML_TARGETS defined in common.mk
# this rule is broken out from common to allow for custom dependencies in the public target
public: $(HTML_TARGETS) sync-svg-colors favicon sync-domain-names

sync-domain-names:
	bash $(MAINT_DIR)/sync-domain-names.sh

# include after public: so public is the default target - provides sync-svg-colors and favicon targets
include $(MAINT_DIR)/favicon.mk

# .md -> .html override rule for writing/ - they share a common template instead of a sibling .template.html per .md file
# $(<) expands to the name of the first dependency (the markdown file)
# $(@) expands to the name of the current target (the .html file being built)
# NOTE: this pattern can be used for any collection of pages under a directory that share a template.
# copy and paste this rule and change the directory name
# e.g. writing/%/index.html: writing/%/index.md -> the/group/%/index.html: the/group/%/index.md the/group/shared.template.html
# and, of course TEMPLATE=the/group/shared.template.html
writing/%/index.html: writing/%/index.md $(ARTICLE_TEMPLATE)
	MD=$(<) TEMPLATE=$(ARTICLE_TEMPLATE) HTML=$(@) $(GEN_HTML_SH)

# generate the table of contents and RSS feed. SITE_AUTHOR, SITE_TOOL and DOMAIN defined in common.mk
index.md: templates/gen-toc-article-list.md always
	$(SITE_TOOL_SH) gen-toc \
	    --base-url https://$(DOMAIN) \
	    --site-desc "Writing by $(SITE_AUTHOR)" \
	    --articles-dir writing \
	    --index-file $(@) \
	    --feed-file $(FEED_XML) \
	    --list-template $(ARTICLE_LIST_TEMPLATE)

# feed is generated in same process that updates ToC in index.md
# make -B feed.xml if it is missing and index.md appears up-to-date
$(FEED_XML): index.md

# block out a new article with a given slug
new-article:
	@test -n "$(SLUG)" || { echo "Usage: make mps-new-article SLUG=my-article-slug" >&2; exit 1; }
	@mkdir -p writing/$(SLUG)
	@cp templates/new-article.md writing/$(SLUG)/index.md
	@echo "Created writing/$(SLUG)/index.md"
