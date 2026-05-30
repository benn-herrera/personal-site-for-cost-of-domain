# shared targets for all domains
# common.mk must be included before this file

.PHONY: serve sync-domain-names common-public content

serve:
	@make -C $(PROJECT_ROOT) $@

sync-domain-names:
	@make -C $(PROJECT_ROOT) $@

include $(MAINT_DIR)/favicon.mk

common-public: $(HTML_TARGETS) sync-svg-colors favicon sync-domain-names

# Deterministic, cross-platform generated content only: HTML, feed.xml + the
# index.md article-list section (byproducts of the HTML_TARGETS build), the
# color-synced favicon.svg, and the src/index.js site list. Deliberately
# EXCLUDES favicon.ico — its qlmanage/Inkscape render is platform-dependent and
# non-reproducible, so it is hand-managed. Called per-site by the pre-commit hook.
content: $(HTML_TARGETS) sync-svg-colors sync-domain-names
