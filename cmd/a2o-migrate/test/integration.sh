#!/bin/sh

RT=${RT:-balena-engine}
CONTAINERIZED=${CONTAINERIZED:-}
PROJECT="$(dirname $(readlink -f $0))/.."
IMAGE=${IMAGE:-balena/balena-engine:beind}

container_name="balena_a2o_integration_test"
balena_container_flags="--rm --detach --name ${container_name} --privileged -v varlibbalena:/var/lib/balena-engine -v ${PROJECT}:/src:ro -w /src"

set -ex

[ -n "$CONTAINERIZED" ] && {

    cat /etc/os-release
    balena-engine info || exit 1

    # ls -lR /var/lib/balena-engine/

    ./a2o-migrate -version
    ./a2o-migrate -debug -migrate

    # ls -lR /var/lib/balena-engine/

    exit 0
}

test_out_dir=$(mktemp -d /tmp/a2o-migrate_test_XXXX)
trap "{ rm -rf ${test_out_dir}; }" EXIT

# start balenaEngine with aufs
$RT run $balena_container_flags $IMAGE --debug --storage-driver=aufs
trap "{ ${RT} stop -t 3 ${container_name}; ${RT} volume rm -f varlibbalena; }" EXIT

sleep 1
$RT exec ${container_name} balena-engine info || exit 1

$RT exec -i ${container_name} balena-engine build -t a2o-test - <<EOF
FROM busybox
RUN mkdir /tmp/d1 && touch /tmp/d1/d1f1 && touch /tmp/f1 && touch /tmp/f2
RUN rm -R /tmp/d1 && mkdir /tmp/d1 && touch /tmp/d1/d1f2 && rm /tmp/f1
RUN ln -s /tmp/d1/d1f2 /tmp/flnk
EOF

$RT exec ${container_name} balena-engine run --name a2o-test-container a2o-test ls -lR /tmp > ${test_out_dir}/stdout_before

# run migration
$RT exec -e CONTAINERIZED=1 ${container_name} /src/test/$(basename $0)

# stop aufs daemon
$RT stop -t 3 ${container_name}

# start balenaEngine with overlay2
$RT run $balena_container_flags $IMAGE --debug --storage-driver=overlay2
sleep 1
$RT inspect ${container_name} &>/dev/null || exit 1

# check if we still are able to create a container from the a2o-test image
$RT exec ${container_name} balena-engine run --rm a2o-test ls -lR /tmp > ${test_out_dir}/stdout_after
# check if rewriting the container storage drivers worked
$RT exec ${container_name} balena-engine start a2o-test-container

# check ls -lR /tmp output
diff ${test_out_dir}/stdout_before ${test_out_dir}/stdout_after
