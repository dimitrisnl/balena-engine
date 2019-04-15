PWD := $(shell pwd)
BIND_IMAGE ?= "balena:bind"

a2o-migrate:
	go build -ldflags='-s' -o ./$@ .

test-unit:
	go test -v ./a2o

test-integration: a2o-migrate
	docker run --rm -it \
		-v $(PWD)/test/integration.sh:/usr/local/bin/test-integration.sh \
		-v $(PWD)/a2o-migrate:/usr/local/bin/a2o-migrate \
		$(BIND_IMAGE) \
		test-integration.sh

legacy:
	docker run --rm -it \
		-v $(PWD)/test/a2o-migrate.sh:/usr/local/bin/a2o-migrate.sh \
		$(BIND_IMAGE) \
		ash -c 'apk add bash && a2o-migrate.sh'

clean:
	$(RM) ./a2o-migrate

all: clean test-unit test-integration

.PHONY: \
	all \
	clean \
	test-unit \
	test-integration
