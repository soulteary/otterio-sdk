# Repository guide

This file records development commands and layout for OtterIO SDK. User-facing setup is in [README.md](./README.md); contribution and live-server test details are in [CONTRIBUTING.md](./CONTRIBUTING.md).

## Commands

Use the Go version declared in `go.mod` (currently Go 1.27.1 or newer).

```sh
# Local checks without a storage server
go test -short -race ./...
go build ./...

# Linter pinned in go.mod; make vet is an alias
make lint

# Compile each independent example program
make examples
```

`core_test.go` skips live-server tests when `SERVER_ENDPOINT` is unset or `-short` is enabled. To run those cases, set `SERVER_ENDPOINT`, `ACCESS_KEY`, `SECRET_KEY`, and `ENABLE_HTTPS`, then run `go test -race ./...`. They create and delete test data; use a disposable server.

The separate functional program has a `mint` build constraint and can be compiled by naming its file:

```sh
go build -race -o /tmp/otterio-sdk-functional-tests functional_tests.go
# Run only after exporting the test connection variables described above.
MINT_MODE=full /tmp/otterio-sdk-functional-tests
```

Without `SERVER_ENDPOINT`, the functional program defaults to the upstream public `play.min.io` endpoint. It covers extensions that are not implemented by every S3 service. See the CI workflows for the current test server setup.

`make checks` includes functional tests. The `make test`, `make functional-test`, and `make functional-test-notls` targets embed upstream localhost test credentials; use explicit commands when testing different credentials or endpoints.

The directories `examples/s3` and `examples/minio` are nested Go modules that replace the SDK dependency with the local checkout. Each file is a separate executable:

```sh
cd examples/s3
go build -o /tmp/otterio-sdk-listbuckets listbuckets.go
```

Configure endpoints, credentials, bucket names, and file paths before running examples. Do not build or run all their `main` functions as one package.

## Layout and API conventions

The root module is `github.com/soulteary/otterio-sdk/v7`; the root Go package remains `minio`.

- `api.go` defines `Client`, `Options`, construction, transport, and request dispatch.
- `api-bucket-*.go` and `api-object-*.go` contain bucket and object operations. `api-get-*.go`, `api-put-*.go`, `api-list.go`, and `api-stat.go` handle reads, writes, listing, and object metadata.
- `pkg/credentials` supplies static, environment, file, IAM, and STS credential providers.
- `pkg/signer` implements S3 request signing. Preserve protocol-required names when editing it.
- `pkg/encrypt`, `pkg/notification`, `pkg/policy`, `pkg/lifecycle`, and `pkg/tags` provide operation-specific helpers.
- Tests live next to implementation files; `functional_tests.go` is a separate live-server suite.
- `docs/API.md` describes exported operations. Availability of an SDK method does not establish support by a particular server.

Network operations use `context.Context` and operation-specific options. Check errors returned when reading lazy `GetObject` readers as well as errors returned when constructing them, and close readers when finished. Use the existing error types and `ToErrorResponse` where S3 error codes matter.
