#!/bin/sh

set -e -o pipefail

RT="podman"
CONTAINERIZED=${CONTAINERIZED:-0}
PROJECT="$(dirname $(readlink -f $0))/../"

set -x

if [[ $CONTAINERIZED -eq 1 ]]; then
    # start balenaEngine
    $RT run --rm --detach --name balena --privileged -v $PROJECT:/src -w /src balena:bind --debug --storage-driver=aufs --storage-opt=sync_diffs=false
    # exec this in the balena container
    $RT exec -it balena /src/test/$(basename $0)
    exit 0
fi

cat /etc/os-release

balena info

echo 'FROM busybox
RUN mkdir /tmp/d1 && touch /tmp/d1/d1f1 && touch /tmp/f1 && touch /tmp/f2
RUN rm -R /tmp/d1 && mkdir /tmp/d1 && touch /tmp/d1/d1f2 && rm /tmp/f1' \ |
    balena build -t a2o-test -

./a2o-migrate -version

./a2o-migrate -debug
