#!/usr/bin/env bash
# site-tool wrapper — resolves the correct binary for dev builds or dist packages.
set -e

TOOL_NAME=$(basename "${0}")
TOOL_NAME=${TOOL_NAME%.sh}
SCRIPT_DIR=$(dirname "${0}")
# realpath is not available on macOS (its bash is too old)
SCRIPT_DIR=$(cd "${SCRIPT_DIR}" && pwd)

OS=$(uname -s)
ARCH=$(uname -m)
EXE=
case "${OS}" in
    Darwin)  OS=darwin  ;;
    Linux)   OS=linux   ;;
    MINGW*|MSYS*|CYGWIN*) OS=windows; EXE=.exe ;;
    *) echo "${TOOL_NAME}: unsupported OS: ${OS}" >&2; exit 1 ;;
esac
case "$ARCH" in
    x86_64)        ARCH=amd64 ;;
    aarch64|arm64) ARCH=arm64 ;;
    *) echo "${TOOL_NAME}: unsupported architecture: ${ARCH}" >&2; exit 1 ;;
esac

is_valid() {
    local bin="$1"
    # silently run version to get an exit code.
    # if there's some weird os version incompatibility
    # with the prebuilt executable this will catch it.
    [[ -x "$bin" ]] && ("$bin" version 1>&2) 2> /dev/null
}

BIN="${SCRIPT_DIR}/bin/${TOOL_NAME}${EXE}"
# Dev build: prefer bin/${TOOL_NAME} (built by `make build`)
if ! is_valid "${BIN}"; then
    # Dist package: use OS and ARCH to construct name ${TOOL_NAME}-OS-ARCH
    BIN="${SCRIPT_DIR}/bin/${TOOL_NAME}-${OS}-${ARCH}${EXE}"
fi

if ! is_valid "${BIN}"; then
    echo "${TOOL_NAME}: no usable binary found." >&2
    echo "  'make -C maint ${TOOL_NAME}' first." >&2
    exit 1
fi

exec "${BIN}" "$@"
