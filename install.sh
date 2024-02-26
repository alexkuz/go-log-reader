#!/bin/sh
set -e

CLR_OFF=""
CLR_RED=""
CLR_GREEN=""
CLR_DIM=""

if [ -t 1 ]; then
    CLR_OFF='\033[0m'
    CLR_RED='\033[0;31m'
    CLR_GREEN='\033[0;32m'
    CLR_DIM='\033[0;2m'
fi

error() {
    echo  "${CLR_RED}error${CLR_OFF}:" "$@" >&2
    exit 1
}

info() {
    echo "${CLR_DIM}" "$@" "${CLR_OFF}"
}

success() {
    echo "${CLR_GREEN}" "$@" "${CLR_OFF}"
}

case $(uname -ms) in
'Darwin x86_64')
    TARGET=Darwin_x86_64
    ;;
'Darwin arm64')
    TARGET=Darwin_arm64
    ;;
'Linux aarch64' | 'Linux arm64')
    TARGET=Linux_arm64
    ;;
'Linux x86_64' | *)
    TARGET=Linux_x86_64
    ;;
esac

INSTALL_DIR=$HOME/.local/bin
ZIP_URI=https://github.com/alexkuz/go-log-reader/releases/latest/download/go-log-reader_$TARGET.tar.gz
TMP_DIR=$(mktemp -d)

info "downloading $ZIP_URI"
curl --progress-bar -L $ZIP_URI | tar xz -C "$TMP_DIR" || error "failed to download $ZIP_URI"

cp -f "$TMP_DIR/go-log-reader" "$INSTALL_DIR" || error "failed to install go-log-reader"

rm -rf "$TMP_DIR"

success "go-log-reader is installed to $INSTALL_DIR"
