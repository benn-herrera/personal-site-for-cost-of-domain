# Shared configuration and common .md -> html rule for all site roots.
# Included by each content root's Makefile.
# Includer must define MAINT_DIR (path to maint/ relative to site dir) before including.

# configure the author name
SITE_AUTHOR  := My Name

# routes execution to the correct bin/site-tool binary for os/architecture
SITE_TOOL_SH := $(MAINT_DIR)/site-tool.sh
# wraps pandoc - detects if not installed, handles filter parsing from frontmatter
GEN_HTML_SH := $(MAINT_DIR)/gen-html.sh

# the content root directory name is the same as the bare domain name (e.g. my-personal-site.me)
DOMAIN := $(notdir $(CURDIR))

# recursively finds all page index.md markdown sources and their corresponding html targets under the current directory
# ignore 'templates' directories
MD_SRCS      := $(shell find . -type f -name '*.md' | grep -v '/templates/')
HTML_TARGETS := $(MD_SRCS:%.md=%.html)

# set to true to use Node.js package 'prettier' to format generated html - see gen-html.sh
export PRETTY_HTML := false

.PHONY: always

# base rule for converting .md to .html
# for any .html file it relies on a .md file and a .template.html file of the same base name
# the yaml frontmatter in the .md file may contain a 'filters' list with pandoc filters to apply
# the filters are specified by bare name (e.g. footnotes) the full path to the lua filter is
# built up in gen-html.sh
# $(<) expands to the name of the first dependency (the markdown file)
# $(word 2,$^) expands to the second dependency (the .template.html pandoc template file)
# $(@) expands to the name of the current target (the .html file being built)
%.html: %.md %.template.html
	MD=$(<) TEMPLATE=$(word 2,$^) HTML=$(@) $(GEN_HTML_SH)
