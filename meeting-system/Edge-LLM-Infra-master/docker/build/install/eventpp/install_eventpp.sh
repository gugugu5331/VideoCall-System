#!/usr/bin/env bash

set -e

cd "$(dirname "${BASH_SOURCE[0]}")"

THREAD_NUM=$(nproc)

VERSION="master"
PKG_NAME="eventpp-${VERSION}.tar.gz"

URL="https://github.com/wqking/eventpp/archive/refs/heads/${VERSION}.tar.gz"

if [ ! -f "${PKG_NAME}" ]; then
    echo "Downloading eventpp (${VERSION})..."
    wget -O "${PKG_NAME}" "${URL}"
fi

tar xzf "${PKG_NAME}"
pushd "eventpp-${VERSION}"
mkdir build && cd build

cmake .. \
    -DCMAKE_INSTALL_PREFIX:PATH="/usr/local"

make -j$(nproc)
make install
ldconfig
popd

rm -rf "${PKG_NAME}" "eventpp-${VERSION}"
