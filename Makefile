GOPATH := $(shell go env GOPATH)
TMPDIR := $(shell mktemp -d)
GOLANGCI_LINT ?= go tool golangci-lint

all: checks

.PHONY: examples docs require-server test functional-test functional-test-notls

checks: lint test examples functional-test

lint:
	@echo "Running $@ check"
	$(GOLANGCI_LINT) run

vet: lint

require-server:
	@: "$${SERVER_ENDPOINT:?Set SERVER_ENDPOINT for a disposable test server}"
	@: "$${ACCESS_KEY:?Set ACCESS_KEY}"
	@: "$${SECRET_KEY:?Set SECRET_KEY}"

test: require-server
	@GO111MODULE=on MINT_MODE=full go test -race -count=1 -v ./...

examples:
	@echo "Building s3 examples"
	@cd ./examples/s3 && $(foreach v,$(wildcard examples/s3/*.go),go build -mod=mod -o ${TMPDIR}/$(basename $(v)) $(notdir $(v)) || exit 1;)
	@echo "Building storage extension examples"
	@cd ./examples/minio && $(foreach v,$(wildcard examples/minio/*.go),go build -mod=mod -o ${TMPDIR}/$(basename $(v)) $(notdir $(v)) || exit 1;)

functional-test: require-server
	@GO111MODULE=on go build -race functional_tests.go
	@MINT_MODE=full ./functional_tests

functional-test-notls: require-server
	@GO111MODULE=on go build -race functional_tests.go
	@ENABLE_HTTPS=0 MINT_MODE=full ./functional_tests

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
