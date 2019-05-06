#!/bin/sh

RT=${RT:-balena-engine}
CONTAINERIZED=${CONTAINERIZED:-0}
PROJECT="$(dirname $(readlink -f $0))/.."
IMAGE=${IMAGE:-balena/balena-engine:beind}

balena_container_flags="--rm --detach --name balena --privileged -v varlibbalena:/var/lib/balena-engine -v ${PROJECT}:/src:ro -w /src"

set -x

[ $CONTAINERIZED -eq 1 ] && {
    # start balenaEngine with aufs
    $RT run $balena_container_flags $IMAGE --debug --storage-driver=aufs
    sleep 1
    $RT inspect balena || exit 1

    # run migration
    $RT exec -it balena /src/test/$(basename $0)
    $RT stop -t 3 balena

    # start balenaEngine with overlay2
    $RT run $balena_container_flags $IMAGE --debug --storage-driver=overlay2
    sleep 1
    $RT inspect balena || exit 1

    # check if we still are able to create a container from the a2o-test image
    $RT exec -it balena balena-engine run --rm a2o-test ls -lR /tmp
    $RT stop -t 3 balena
    $RT volume rm -f varlibbalena
    exit 0
}

cat /etc/os-release

balena-engine info || exit 1

cat test/Dockerfile | balena-engine build -t a2o-test -

balena-engine run --rm a2o-test ls -lR /tmp

ls -l /var/lib/balena-engine/aufs/
ls -l /var/lib/balena-engine/overlay2/

./a2o-migrate -version

./a2o-migrate -debug

ls -l /var/lib/balena-engine/aufs/
ls -l /var/lib/balena-engine/overlay2/
