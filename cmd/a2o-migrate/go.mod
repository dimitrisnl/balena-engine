module github.com/balena-os/balena-engine/cmd/a2o-migrate

go 1.12

require (
	github.com/docker/docker/daemon/graphdriver/overlay2 v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.4.1
	golang.org/x/sys v0.0.0-20190506115046-ca7f33d4116e
	golang.org/x/xerrors v0.0.0-20190410155217-1f06c39b4373
)

replace github.com/docker/docker/daemon/graphdriver/overlay2 => ./vendor/overlay2
