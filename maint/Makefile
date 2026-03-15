# Run from repo root: make -C maint [target]
# Or: cd maint && make [target]
#
# Site build targets live in each site's directory:
#   cd public/<domain-name>; make [target]
#
# Targets here:
#   serve           — start wrangler dev server on 0.0.0.0
#   site-tool       — build site-tool binary for current platform
#   site-tool-dist  — cross-compile site-tool for all platforms
#   site-tool-clean — remove maint/bin/

.PHONY: serve site-tool site-tool-dist site-tool-clean site-tool-nuke sync-domain-names

serve:
	@echo "specify simulated domain with ?d=[site-domain.tld]"
	cd .. && wrangler dev --ip=0.0.0.0

site-tool:
	$(MAKE) -C site-tool build

site-tool-clean:
	$(MAKE) -C site-tool clean

site-tool-dist:
	$(MAKE) -C site-tool dist

site-tool-nuke:
	$(MAKE) -C site-tool nuke

sync-domain-names:
	@bash sync-domain-names.sh
