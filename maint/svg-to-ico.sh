#!/usr/bin/env bash
# svg-to-ico.sh — convert SVG to ICO
# macOS - uses qlmanage + ImageMagick
# other platforms uses ImageMagick directly and warns to check output
#
# Usage: svg-to-ico.sh <input.svg> <output.ico>

set -e

SVG="$1"
ICO="$2"

if [ -z "${SVG}" ] || [ -z "${ICO}" ]; then
    echo "Usage: $0 <input.svg> <output.ico>" >&2
    exit 1
fi

OS_NAME=$(uname -s)
MAGICK=$(which magick 2> /dev/null)

if [[ ! -x "${MAGICK}" ]]; then
    echo "svg-to-ico: ImageMagick not found in PATH." 1>&2
    if [[ "${OS_NAME}" = "Darwin" ]]; then
        echo "'brew install imagemagick' to install"
    elif [[ "${OS_NAME}" = "Linux" ]]; then
        echo "'sudo apt install imagemagick' to install"
    else
        echo "winget install -e ImageMagick.ImageMagick"
    fi
    exit 1
fi

if [[ "${OS_NAME}" = "Darwin" ]]; then
    # qlmanage SVG renderer is complete and reliable.
    qlmanage -t -s 512 -o /tmp/ "${SVG}" 2>/dev/null
    magick /tmp/favicon.svg.png -define icon:auto-resize=32,16 "${ICO}"
else
    # ImageMagick SVG renderer is not fully complete. YMMV.
    magick "${SVG}" -define icon:auto-resize=32,16 "${ICO}"
    echo "on ${OS_NAME} ImageMagick's incomplete SVG renderer is used directly. Verify your results."
    echo "if the output is bad, try using a free online tool to do the conversion as a one-off."
fi
