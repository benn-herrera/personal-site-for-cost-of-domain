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

.PHONY: serve sync-domain-names site-tool site-tool-dist site-tool-clean site-tool-nuke

# also available from the public/<domain-name> Makefile
serve:
	@echo "specify simulated domain with ?d=[site-domain.tld]"
	wrangler dev --ip=0.0.0.0

# also available from the public/<domain-name> Makefile
sync-domain-names:
		@bash ./maint/sync-domain-names.sh

# targets below are only available in the top level Makefile

site-tool:
	$(MAKE) -C ./maint/site-tool build

site-tool-clean:
	$(MAKE) -C ./maint/site-tool clean

site-tool-dist:
	$(MAKE) -C ./maint/site-tool dist

site-tool-nuke:
	$(MAKE) -C ./maint/site-tool nuke
