#!/bin/bash

set -ex

arch=${BUILD_ARCH:-$(uname -m)}
version=${VERSION:-$(git describe --tags --always)}
target=dynbinary-balena

# build env defaults
export GOMAXPROCS=1
export VERSION="$version"
export DOCKER_LDFLAGS="-s" # strip resulting binary

# overwrite defaults for a static build
if [ -n "${BUILD_STATIC}" ]; then
    target=binary-balena
    version="$version-static"
fi

# overwrite defaults when targeting balenaOS
if [ -n "${BUILD_BALENAOS}" ]; then
    export DOCKER_BUILDTAGS='exclude_graphdriver_btrfs exclude_graphdirver_zfs exclude_graphdriver_devicemapper no_btrfs'
    version="$version-balenaos"
fi

src="bundles/latest/$target"
dst="balena-engine"

# run the build
(
    rm -rf "$src/*" || true
    ./hack/make.sh "$target"
)

# pack the release artifacts
(
    rm -rf "$dst" || true
    mkdir "$dst"
    cp --no-dereference "$src"/balena-engine* "$dst/"
    tar czfv "balena-engine-$version-$arch.tar.gz" "$dst"
)

