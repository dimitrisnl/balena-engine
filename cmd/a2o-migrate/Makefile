PWD := $(shell pwd)

a2o-migrate: clean
	go build -ldflags="-s -X main.GitVersion=$(shell git describe --tag --always) -X main.BuildTime=$(shell date +%FT%T)"  -o ./$@ .

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
