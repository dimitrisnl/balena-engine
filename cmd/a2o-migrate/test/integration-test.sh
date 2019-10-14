#!/bin/sh

cat /etc/os-release
balena-engine info || exit 1

# ls -lR /var/lib/balena-engine/

./a2o-migrate -version
./a2o-migrate -debug -migrate
./a2o-migrate -debug -commit

# ls -lR /var/lib/balena-engine/
