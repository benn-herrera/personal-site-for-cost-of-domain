#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(dirname "${0}")
SCRIPT_DIR="$(cd "${SCRIPT_DIR}" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
INDEX_JS="${REPO_ROOT}/src/index.js"
JS_TEXT=$(cat "${INDEX_JS}")
SITES_DECL="    const sites = "

# Collect sorted directory names from public/
domains=$(cd "${REPO_ROOT}/public" && find . -maxdepth 1 -mindepth 1 -type d | sed 's:^\./::' | sort | tr '\n' ' ')
# trim trailing space and quote each domain
domains=${domains% }
domains=\"${domains// /\", \"}\"

new_sites_line="    const sites = [${domains}];"
cur_sites_line=$(echo "${JS_TEXT}" | grep '    const sites = \[')

# Compare to existing; skip write if unchanged (preserve timestamp)
[[ "${cur_sites_line}" == "${new_sites_line}" ]] && exit 0

# Replace in-place with awk (portable: no sed -i dialect differences)
echo "${JS_TEXT}" | awk -v NEW_SITES_LINE="${new_sites_line}" '
    /^    const sites = \[/ { print NEW_SITES_LINE; next; }
    /.*/ { print; }
    ' > "${INDEX_JS}"

echo "sync-domain-names: src/index.js updated with site list [${domains}]"
