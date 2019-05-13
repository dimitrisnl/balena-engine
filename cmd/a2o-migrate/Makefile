PWD := $(shell pwd)

GO_BUILDINFO=-X main.GitVersion=$(shell git describe --tag --always) -X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_LDFLAGS=-extldflags "-fno-PIC -static"
GO_BUILDFLAGS=-ldflags '$(GO_LDFLAGS) $(GO_BUILDINFO)' -buildmode pie -tags 'osusergo netgo static_build'

a2o-migrate: clean
	go build $(GO_BUILDFLAGS) -o ./$@ .

test-unit:
	go test -v ./a2o

test-integration: a2o-migrate
	env CONTAINERIZED=1 ./test/integration.sh

legacy:
	$(RUNTIME) run --rm -it \
		-v $(PWD)/test/a2o-migrate.sh:/usr/local/bin/a2o-migrate.sh \
		$(BEIND_IMAGE) \
		ash -c 'apk add bash && a2o-migrate.sh'

clean:
	$(RM) ./a2o-migrate

all: clean test-unit test-integration

.PHONY: \
	all \
	clean \
	test-unit \
	test-integration
