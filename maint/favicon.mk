# shared targets for favicon.svg color sync and .ico generation from SVG
# common.mk must be included before this file

CSS         := style.css
FAVICON_SVG := favicon.svg
FAVICON_ICO := favicon.ico

.PHONY: sync-svg-colors favicon

sync-svg-colors: $(FAVICON_SVG)

# updates the light/dark mode color sets in favicon.svg
# favicon can't rely on CSS directly
$(FAVICON_SVG): $(CSS)
	$(SITE_TOOL_SH) sync-svg-colors --css $(CSS) $@

favicon: $(FAVICON_ICO)

# creates a static .ico from the favicon.svg as a fallback for Safari, which doesn't support SVG favicons
$(FAVICON_ICO): $(FAVICON_SVG)
	$(MAINT_DIR)/svg-to-ico.sh $(abspath $(FAVICON_SVG)) $@
