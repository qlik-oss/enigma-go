![](enigma-go.png)

[![validate](https://github.com/qlik-oss/enigma-go/actions/workflows/validate.yml/badge.svg)](https://github.com/qlik-oss/enigma-go/actions/workflows/validate.yml)
[![Go Report Card](https://goreportcard.com/badge/qlik-oss/enigma-go)](https://goreportcard.com/report/qlik-oss/enigma-go)
![Latest Version](https://img.shields.io/github/release/qlik-oss/enigma-go.svg?style=flat)

enigma-go is a library that helps you communicate with a Qlik Associative Engine.
Examples of use may be building your own analytics tools, back-end services, or other tools communicating with a Qlik Associative Engine.

---

- [Installation](#installation)
- [Getting started](#getting-started)
- [Schemas](#schemas)
- [Contributing and commit messages](#contributing-and-commit-messages)
- [Release](#release)
- [Updating enigma-go to new engine schema](#updating-enigma-go-to-new-engine-schema)

---

## Installation

```bash
go get -u github.com/qlik-oss/enigma-go/v4
```

## Getting started

Connecting to a Qlik Associative Engine and interacting with a document/app involves at least the following steps:

1. Create and set up a Dialer object with TLS configuration, etc.

2. Open a WebSocket to the Qlik Associative Engine using the Dial function in the Dialer.

3. Open or create a document/app using openDoc or createApp.

Refer to the [examples](https://github.com/qlik-oss/enigma-go/tree/master/examples) section for more information.

## Schemas

enigma-go includes generated API code that is based on the latest available Qlik Associative Engine schema.
When a new schema is available, a new version of enigma-go will be made available.

## Contributing and commit messages

This project uses [conventional commits](https://www.conventionalcommits.org), for example `feat: add X`,
`fix(deps): update Y` or `feat!: remove Z` for breaking changes. Pull request titles are validated by the
[Semantic PR](./.github/workflows/semantic.yaml) workflow, and the commit messages are used to determine the
version of the next release. See also [CONTRIBUTING](./.github/CONTRIBUTING.md).

## Release

Releases are created by the [release-enigma-go](./.github/workflows/release-enigma-go.yml) GitHub workflow, which is run
manually (`workflow_dispatch`) on the **master** branch. It takes two inputs:
- `version`: `auto`, `major`, `minor` or `patch`. With `auto` the bump is derived from the conventional commits since the
previous tag: a breaking change (`!` or a `BREAKING CHANGE` footer) gives `major`, `feat` gives `minor` and `fix`/`perf`
gives `patch`. If there are none of these the release is aborted.
- `dry_run`: shows what would be pushed without pushing anything.

The workflow runs the [release.sh](./release/release.sh) script, pushes the resulting tag and creates a GitHub release with
auto-generated release notes.

The script can also be run locally from the **master** branch:
```bash
./release/release.sh <major|minor|patch|auto>
```
It does the following:
1. Checks that the local repo is pristine (no untracked files or uncommitted changes), that you are on **master** and
that you are up-to-date with the remote **master**.
2. Creates a new version based on the previous version-tag (if any, otherwise 0.0.0) and the requested bump.
3. Creates an annotated tag `v<version>` with the message `Release: v<version> for QIX schema version <qix_version>`.

The script does not push anything. If everything looks good, run:
```bash
git push --follow-tags
```

## Updating enigma-go to new engine schema

- Generate enigma-go: [Generate-Docs](./schema/README.md)
