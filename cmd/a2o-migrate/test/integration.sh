#!/bin/sh

RT=${RT:-docker}
CONTAINERIZED=${CONTAINERIZED:-}
PROJECT="$(dirname $(readlink -f $0))/.."
IMAGE=${IMAGE:-balena/balena-engine:beind}

test_out_dir=$(mktemp -u --tmpdir -d a2o-migrate_test.XXXX)
balena_container_flags="--rm --detach --name balena --privileged -v varlibbalena:/var/lib/balena-engine -v ${PROJECT}:/src:ro -w /src"

set -ex

[ -n "$CONTAINERIZED" ] && {

    cat /etc/os-release
    balena-engine info || exit 1

    # ls -lR /var/lib/balena-engine/

    ./a2o-migrate -version

    ./a2o-migrate -debug -migrate

    cat /lib/systemd/system/balena.service | grep overlay2
    cat /etc/systemd/system/balena.service.d/balena.conf | grep overlay2

    # ls -lR /var/lib/balena-engine/

    exit 0
}

# start balenaEngine with aufs
$RT run $balena_container_flags $IMAGE --debug --storage-driver=aufs
sleep 1
$RT exec balena balena-engine info || exit 1

$RT exec -i balena balena-engine build -t a2o-test - <<EOF
FROM busybox
RUN mkdir /tmp/d1 && touch /tmp/d1/d1f1 && touch /tmp/f1 && touch /tmp/f2
RUN rm -R /tmp/d1 && mkdir /tmp/d1 && touch /tmp/d1/d1f2 && rm /tmp/f1
EOF

$RT exec balena balena-engine run --name a2o-test-container a2o-test ls -lR /tmp > ${test_out_dir}/stdout_before

# copy systemd files into container
$RT exec balena ash -c 'mkdir -p /lib/systemd/system ; mkdir -p /etc/systemd/system'
$RT cp $PROJECT/test/systemd/balena.service balena:/lib/systemd/system/balena.service
$RT cp $PROJECT/test/systemd/balena.service.d balena:/etc/systemd/system/balena.service.d

# run migration
$RT exec -it -e CONTAINERIZED=1 balena /src/test/$(basename $0)

# check if we can still run from the aufs image
$RT exec balena balena-engine run --rm a2o-test ls -lR /tmp
# stop aufs daemon
$RT stop -t 3 balena

# start balenaEngine with overlay2
$RT run $balena_container_flags $IMAGE --debug --storage-driver=overlay2
sleep 1
$RT inspect balena &>/dev/null || exit 1

# check if we still are able to create a container from the a2o-test image
$RT exec balena balena-engine run --rm a2o-test ls -lR /tmp > ${test_out_dir}/stdout_after
# check if rewriting the container storage drivers worked
$RT exec balena balena-engine start a2o-test-container

# check if ls -lR /tmp returned the same in the aufs and overlay2 containers
diff -q ${test_out_dir}/stdout_before ${test_out_dir}/stdout_after || exit 1

# cleanup
$RT stop -t 3 balena
$RT volume rm -f varlibbalena
