# Contributing to OtterIO SDK

Submit issues and pull requests to [soulteary/otterio-sdk](https://github.com/soulteary/otterio-sdk). Include the SDK revision, Go version, target server and version, and a small reproduction when reporting a problem. Remove credentials and private data from examples and logs.

## Local development

Use Go 1.27.2 or newer, as declared in [go.mod](./go.mod). Fork or clone the repository, create a branch, and make a focused change. The SDK module and its helper imports use `github.com/soulteary/otterio-sdk/v7`; the root Go package is still named `minio`.

The `go.mod` tool directive pins golangci-lint. Keep its transitive dependencies compatible: golangci-lint v2.14.0 still requires `gobwas/glob v0.2.3`, `nishanths/exhaustive v0.13.0`, and `nishanths/predeclared v0.2.2`; newer versions remove APIs used by the linter. Run `make lint` after updating tool dependencies.

For code changes, format the edited Go files and add tests that cover the changed behavior. Follow [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments). Keep original copyright and license notices.

Run these checks from the repository root:

```sh
go test -short -race ./...
go build ./...
make lint
make examples
```

`make lint` uses the `golangci-lint` tool pinned in `go.mod` through `go tool golangci-lint`; `make vet` is an alias for the same check. `make examples` builds every example file separately because each has its own `main` function. The example directories have their own `go.mod` files and replace the SDK dependency with the local checkout.

For documentation changes, check relative links and anchors and compile any new runnable Go examples against this checkout. Describe what you verified in the pull request; no new test is needed for a prose-only correction.

## Tests with a live server

`go test -short -race ./...` skips the live-server tests in `core_test.go`. Those tests also skip when `SERVER_ENDPOINT` is unset. Other tests use local fixtures or mock HTTP servers and do not require an AWS account.

To include the live-server tests, configure a **disposable test server** and test credentials, then run:

```sh
: "${SERVER_ENDPOINT:?Set the test S3 endpoint, for example 127.0.0.1:9000}"
: "${ACCESS_KEY:?Set the test access key}"
: "${SECRET_KEY:?Set the test secret key}"
: "${ENABLE_HTTPS:?Set true for HTTPS or false for HTTP}"
export SERVER_ENDPOINT ACCESS_KEY SECRET_KEY ENABLE_HTTPS
go test -race ./...
```

These tests create and remove buckets and objects. Use the S3 API endpoint rather than a separate web console port. For a private test CA, configure `SSL_CERT_FILE` with the trusted CA certificate.

The larger functional suite is a separate program in [functional_tests.go](./functional_tests.go), excluded from normal package builds by the `mint` build constraint. With the environment above set, compile the file directly and run it:

```sh
go build -race -o /tmp/otterio-sdk-functional-tests functional_tests.go
MINT_MODE=full /tmp/otterio-sdk-functional-tests
```

Do not run that program without `SERVER_ENDPOINT`: it exits before making requests when the endpoint or credentials are absent. The full functional suite contains AWS- and MinIO/AIStor-specific cases; a server that implements basic S3 operations may not implement every case. Default Linux and Windows CI run local tests and eight core operation tests against a pinned OtterIO server. Extended AIStor coverage is a manual workflow requiring a reviewed image digest and the `AISTOR_LICENSE` secret. The [Linux](./.github/workflows/go.yml) and [Windows](./.github/workflows/go-windows.yml) workflows show the current CI server and configuration, including encryption and other features needed by the suite.

`make checks` combines lint, package tests, example builds, and functional tests. Its `test` and `functional-test` targets hard-code a TLS server at `localhost:9000` with upstream test credentials, so use the explicit commands above for a different server or credentials.

## Pull requests

Explain the problem, the resulting behavior, and your verification. Link a related issue when one exists. Keep README commands, API examples, and the [Chinese quick start](./README_zh_CN.md) consistent when changing user-facing behavior. Release preparation is described in [MAINTAINERS.md](./MAINTAINERS.md).
