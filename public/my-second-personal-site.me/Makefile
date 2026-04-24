MAINT_DIR := ../../maint

# must come first
include $(MAINT_DIR)/common.mk

.PHONY: public

# common_public and serve defined in common-targets.mk (i.e. 'make serve' works from this directory)
# public is broken out like this so you can add custom dependencies to the default target after common_public
public: common-public

# include after public: so public is the default target - provides serve sync-domain-names sync-svg-colors and favicon targets
include $(MAINT_DIR)/common-targets.mk

# custom rules and targets below this line
