![](enigma-go.png)

[![CircleCI](https://circleci.com/gh/qlik-oss/enigma-go.svg?style=shield)](https://circleci.com/gh/qlik-oss/enigma-go)
[![Go Report Card](https://goreportcard.com/badge/qlik-oss/enigma-go)](https://goreportcard.com/report/qlik-oss/enigma-go)

enigma-go is a library that helps you communicate with a Qlik Associative Engine.
Examples of use may be building your own analytics tools, back-end services, or other tools communicating with a Qlik Associative Engine. As an example Qlik Core provides an easy way to get started.

---

- [Getting started](#getting-started)
- [Qlik Core](https://core.qlik.com/)
- [API documentation](https://godoc.org/github.com/qlik-oss/enigma-go)
- [Contributing](./.github/CONTRIBUTING.md#contributing-to-enigma-go)
- [Runnable examples](./examples/README.md)
- [Generating from schema](./schema/README.md)

---

## Installation

```bash
go get -u github.com/qlik-oss/enigma-go
```

## Getting started

Connecting to a Qlik Associative Engine (e.g Qlik Core) and interacting with a document/app involves at least the following steps:

1. Create and set up a Dialer object with TLS configuration, etc.

2. Open a WebSocket to the Qlik Associative Engine using the Dial function in the Dialer.

3. Open or create a document/app using openDoc or createApp.

Refer to the [examples](https://github.com/qlik-oss/enigma-go/tree/master/examples) section for more information.

## Schemas

enigma-go includes generated API code that is based on the latest available Qlik Associative Engine schema.
When a new schema is available, a new version of enigma-go will be made available.

## Release

To release a new version of enigma-go you have to be on the **master** branch.
From there you can run the [release.sh](./release/release.sh) script. The usage is:
```bash
./release.sh <major|minor|patch>
```
where the argument specifies what should be bumped. The release-script does a couple of things.
1. Creates a new version based on previous version-tag (if any, otherwise 0.0.0) and suffixes it with the QIX schema version
as metadata. For example bumping minor when there are no previous tags will result in the version `0.1.0+12.429.0`.
2. Generates a new API specification using the new version.
3. Adds the resulting `api-spec.json` file to a commit with the message `Release: <version> for QIX schema version <qix_version>`.
4. Creates a tag containing the version with the same message as in step 3.
5. Afterwards, adds another commit bumping the `api-spec.json` to latest again.

After the script has run, check the results. If everything looks good run:
```bash
git push --follow-tags
```
to push the tag and commit to **master**.
The release-script also checks if the local repo is in a pristine state: no untracked files or uncommitted change and, you
have to be up-to-date with the latest changes on **master**.
