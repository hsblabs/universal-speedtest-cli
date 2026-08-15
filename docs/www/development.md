---
schema_version: "2026-08-15"
okf_version: "0.2"
type: guide
title: Building, testing, and releasing
description: How to build unispeedtest from source, run its test suite, and cut a release through Conventional Commits and release-please.
resource: https://github.com/hsblabs/universal-speedtest-cli
tags:
  - go
  - contributing
  - release
status: stable
hsblabs:
  sidebar:
    label: Development
    order: 4
---

This page is for working on `unispeedtest` itself. To install and use the released binary, start at the [overview](./index.md).

## Test

```sh
go test ./...
```

## Build

```sh
go build -trimpath -ldflags="-s -w" -o dist/unispeedtest ./cmd/unispeedtest
```

A build without an embedded version reports `dev` for `--version`; released binaries carry the tag from the module build info.

## Release

Releases are automated and driven by commit messages.

1. Merged PRs follow [Conventional Commits](https://www.conventionalcommits.org/) — `fix:`, `feat:`, and `feat!:` or a `BREAKING CHANGE:` footer for majors.
2. A push to `main` opens or updates a release PR via `release-please`.
3. Merging that release PR updates `version.txt` and `CHANGELOG.md`.
4. The release workflow tags `vX.Y.Z` and publishes artifacts with GoReleaser.

Nothing is tagged or published by hand; the release PR is the only gate.

## License

MIT.
