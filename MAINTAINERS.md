# OtterIO SDK maintainer notes

These notes apply to [soulteary/otterio-sdk](https://github.com/soulteary/otterio-sdk). Releases of this fork use its repository and maintainer credentials; the upstream MinIO signing keys, accounts, and release pages are not part of this process.

## Prepare a release

1. Review the intended revision and the changes since the previous [tag](https://github.com/soulteary/otterio-sdk/tags). Run the checks described in [CONTRIBUTING.md](./CONTRIBUTING.md), including the live-server cases needed by the changes.
2. Confirm that `go.mod`, helper imports, example modules, and documentation use the `github.com/soulteary/otterio-sdk/v7` module path. Tags for this major version must use the `v7.<minor>.<patch>` form.
3. Review `libraryVersion` in [api.go](./api.go). It is part of the inherited client user agent; change it deliberately when preparing a version update, without replacing upstream copyright notices or protocol identifiers.
4. Create the reviewed tag on the intended commit and push it to this fork using your own authorized account. Sign the tag with your own signing key if you use signed tags.
5. Publish notes on this repository's [releases page](https://github.com/soulteary/otterio-sdk/releases), describing changes, compatibility implications, and validation. Identify any upstream-derived changes and link their sources.

The checked-in [workflows](./.github/workflows) run pull request checks and vulnerability scanning; there is currently no release publishing workflow. A tag push does not automatically create a GitHub release or binary assets. This repository provides a Go module, so consumers select a release with `go get github.com/soulteary/otterio-sdk/v7@<tag>`.

## Attribution

Use your own name and contact information for commits. Keep [LICENSE](./LICENSE), [NOTICE](./NOTICE), and original source copyright headers intact. Do not attribute fork releases or commits to an upstream MinIO account.
