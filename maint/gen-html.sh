#!/usr/bin/env bash
# do not compute absolute path - it might contain spaces.
THIS_DIR=$(dirname "${0}")
FILTER_DIR="${THIS_DIR}/filters"

# invocation is with command envar syntax - effectively named parameters, no CLI parsing needed
function usage() {
    echo "Usage: MD=<src.md> HTML=<out.md> TEMPLATE=<src.template.html> gen-html.sh"
    exit 1
}

if [[ -z "${MD}" || -z "${HTML}" || -z "${TEMPLATE}" ]]; then
    usage
fi

if !(2>&1 which pandoc) > /dev/null; then
    echo "pandoc not installed or not in PATH"
    exit 1
fi

function parse_filters() {
    awk -v FILTER_DIR="${FILTER_DIR}/" '
      /^---$/ {
        # detect frontmatter block
        if (in_yaml == 1) { exit(0); }
        in_yaml = 1;
        next;
      }
      /.*/ {
        # ignore everything outside frontmatter block and skip comments/empty lines
        if (in_yaml != 1 || $1 == "#" || $1 == "") { next; }

        # detect start and end of filters list
        if ($1 == "filters:") { in_filters=1; next; }
        if (in_filters != 1) { next; }
        if ($1 != "-") { exit(0); }

        # accumulate filters
        filters=filters " --lua-filter " FILTER_DIR $2 ".lua";
      }
      END {
        print filters;
      }' "${1}"
}

# parse filters list out of yaml frontmatter from markdown
FILTERS=$(parse_filters "${MD}")

PRETTY_HTML=${PRETTY_HTML:-false}

function pretty_print() {
    if ${PRETTY_HTML}; then
        npx prettier --write --parser html "${1}"
    else
        cat > "${1}"
    fi
}

set -x
pandoc "${MD}" --standalone --strip-comments --template "${TEMPLATE}" ${FILTERS} "${@}" | \
    pretty_print "${HTML}"
