MAINT_DIR := ../../maint
# must come first
include $(MAINT_DIR)/common.mk

.PHONY: public sync-domain-names

# HTML_TARGETS defined in common.mk
# this rule is broken out from common to allow for custom dependencies in the public target
public: $(HTML_TARGETS) sync-svg-colors favicon sync-domain-names

sync-domain-names:
	bash $(MAINT_DIR)/sync-domain-names.sh

# include after public: so public is the default target - provides sync-svg-colors and favicon targets
include $(MAINT_DIR)/favicon.mk
