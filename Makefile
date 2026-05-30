# Infra targets — run from the repo root: make [target]
#
# Site build targets live in each site's directory:
#   cd public/<domain-name>; make [target]
#
# Targets here:
#   serve           — start wrangler dev server on 0.0.0.0
#   verify          — run the pre-commit check now (build + stale-content, against staged content)
#   install-hooks   — install the pre-commit hook (builds + stale-content check) for this clone
#   uninstall-hooks — remove the pre-commit hook for this clone
#   site-tool       — build site-tool binary for current platform
#   site-tool-test  — run site-tool unit tests
#   site-tool-dist  — cross-compile site-tool for all platforms
#   site-tool-clean — remove maint/bin/
#
# `verify` is just a name for running maint/githooks/pre-commit by hand, so the
# manual check and the commit-time check are identical. The actual logic (stash
# unstaged work, build, diff) lives in the hook because it's tree-aware; the
# Makefile only offers a way to run it before attempting a commit.

.PHONY: help serve verify install-hooks uninstall-hooks sync-domain-names site-tool site-tool-test site-tool-dist site-tool-clean site-tool-nuke

help:
	@echo "available targets: serve, verify, install-hooks, uninstall-hooks, sync-domain-names, site-tool, site-tool-test, site-tool-clean, site-tool-dist, site-tool-nuke"

# also available from the public/<domain-name> Makefile
serve:
	@echo "specify simulated domain with ?d=[site-domain.tld]"
	wrangler dev --ip=0.0.0.0

# also available from the public/<domain-name> Makefile
sync-domain-names:
	@./maint/sync-domain-names.sh

# targets below are only available in the top level Makefile

# Run the pre-commit check on demand: same build + stale-content check the commit
# runs, against the STAGED snapshot (unstaged work is stashed and restored by the
# hook). Lets you check before attempting a commit instead of finding out at commit
# time. Single source of truth — this just invokes the hook.
verify:
	@./maint/githooks/pre-commit

# Point git at the versioned hooks in maint/githooks (the pre-commit hook builds
# the staged snapshot and blocks on stale generated content).
# One-time per clone. core.hooksPath is resolved relative to the repo root.
install-hooks:
	git config core.hooksPath maint/githooks
	@echo "installed: pre-commit hook active (core.hooksPath = maint/githooks)"

# Revert install-hooks: restore git's default hook path (.git/hooks). Idempotent.
uninstall-hooks:
	@git config --unset core.hooksPath 2>/dev/null || true
	@echo "uninstalled: core.hooksPath cleared (git uses .git/hooks again)"

site-tool:
	$(MAKE) -C ./maint/site-tool build

site-tool-test:
	$(MAKE) -C ./maint/site-tool test

site-tool-clean:
	$(MAKE) -C ./maint/site-tool clean

site-tool-dist:
	$(MAKE) -C ./maint/site-tool dist

site-tool-nuke:
	$(MAKE) -C ./maint/site-tool nuke
