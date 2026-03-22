#!/usr/bin/env bash
# svg-to-ico.sh — convert SVG to ICO
# macOS - uses qlmanage + ImageMagick
# other platforms uses ImageMagick directly and warns to check output
#
# Usage: svg-to-ico.sh <input.svg> <output.ico>


SVG="${1}"
ICO="${2}"

if [[ -z "${SVG}" || -z "${ICO}" ]]; then
    echo "Usage: $0 <input.svg> <output.ico>" 1>&2
    exit 1
fi

OS_NAME=$(uname -s)

if [[ -n "${PROGRAMFILES}" ]]; then
    WIN_PF=${PROGRAMFILES//\\//}
    WIN_PF=/${WIN_PF/:/}
    MAGICK_DIR=$(cd "${WIN_PF}"; echo ImageMagick-* | sort -Vr | head -1)
    PATH="${PATH}:${WIN_PF}/Inkscape/bin:${WIN_PF}/${MAGICK_DIR}"
fi

MAGICK=$(which magick 2> /dev/null)
if [[ ! -x "${MAGICK}" ]]; then
    MAGICK=$(which convert 2> /dev/null)
fi

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

TMP_DIR="/tmp/svg-to-ico-${$}"

function rm_tmp_dir() {
    /bin/rm -rf "${TMP_DIR}"
}

trap rm_tmp_dir exit
mkdir "${TMP_DIR}"

TMP_PNG="${TMP_DIR}/tmp.png"

if [[ "${OS_NAME}" = "Darwin" ]]; then
    qlmanage -t -s 512 -o "${TMP_PNG}" "${SVG}" 2>/dev/null
else
    if !(2>&1 which inkscape) > /dev/null; then
        echo "svg-to-ico: Inkscape not installed or not in PATH"
        exit 1
    fi
    inkscape "${SVG}" --export-type=png --export-filename="${TMP_PNG}" -w 512 2> /dev/null
fi

"${MAGICK}" "${TMP_PNG}" -define icon:auto-resize=32,16 "${ICO}"
