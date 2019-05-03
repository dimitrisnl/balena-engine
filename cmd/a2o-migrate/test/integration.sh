#!/bin/sh

RT=${RT:-balena-engine}
CONTAINERIZED=${CONTAINERIZED:-0}
PROJECT="$(dirname $(readlink -f $0))/../"
IMAGE=${IMAGE:-balena/balena-engine:beind}

set -x

[ $CONTAINERIZED -eq 1 ] && {
    # start balenaEngine
    $RT run --rm --detach --name balena --privileged -v $PROJECT:/src -w /src $IMAGE --debug --storage-driver=aufs
    sleep 1
    $RT inspect balena || exit 1
    # exec this in the balena container
    $RT exec -it balena /src/test/$(basename $0)
    $RT stop balena
    exit 0
}

cat /etc/os-release

balena-engine info || exit 1

cat test/Dockerfile | balena-engine build -t a2o-test -

ls -l /var/lib/balena-engine/aufs/
ls -l /var/lib/balena-engine/overlay2/

./a2o-migrate -version

./a2o-migrate -debug
