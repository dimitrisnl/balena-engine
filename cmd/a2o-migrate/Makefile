PWD := $(shell pwd)
VERSION ?= $(shell git describe --tag --always)

GO_BUILDINFO=-X main.GitVersion=$(VERSION) -X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_LDFLAGS=-extldflags "-fno-PIC -static"
GO_BUILDFLAGS=-ldflags '$(GO_LDFLAGS) $(GO_BUILDINFO)' -tags 'osusergo netgo static_build'

a2o-migrate: clean
	go build $(GO_BUILDFLAGS) -o ./$@ ./cli

test-unit:
	go test -cover -v ./...

test-integration: a2o-migrate
	./test/integration.sh

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
