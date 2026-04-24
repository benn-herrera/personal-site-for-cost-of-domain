# shared targets for all domains
# common.mk must be included before this file

.PHONY: serve sync-domain-names common-public

serve:
	@make -C $(PROJECT_ROOT) $@

sync-domain-names:
	@make -C $(PROJECT_ROOT) $@

include $(MAINT_DIR)/favicon.mk

common-public: $(HTML_TARGETS) sync-svg-colors favicon sync-domain-names
