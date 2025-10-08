#!/usr/bin/env bash

set -e

cd "$(dirname "${BASH_SOURCE[0]}")"

THREAD_NUM=$(nproc)


VERSION="master"
PKG_NAME="simdjson-${VERSION}.tar.gz"

tar xzf "${PKG_NAME}"
pushd simdjson-${VERSION}
mkdir build && cd build

cmake .. \
    -DCMAKE_INSTALL_PREFIX:PATH="/usr/local" \
    -DCMAKE_BUILD_TYPE=Release

make -j$(nproc)
make install
ldconfig
popd

rm -rf PKG_NAME simdjson-${VERSION}