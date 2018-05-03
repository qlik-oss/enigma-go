![](enigma-go.png)

[![CircleCI](https://circleci.com/gh/qlik-oss/enigma-go.svg?style=shield)](https://circleci.com/gh/qlik-oss/enigma-go)
[![Go Report Card](https://goreportcard.com/badge/qlik-oss/enigma-go)](https://goreportcard.com/report/qlik-oss/enigma-go)

enigma-go is a library that helps you communicate with a Qlik Associative Engine.
Examples of use may be building your own analytics tools, back-end services, or other tools communicating with a Qlik Associative Engine.

---

- [Getting started](#getting-started)
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

Connecting to a Qlik Associative Engine and interacting with a document/app involves at least the following steps:

1. Create and set up a Dialer object with TLS configuration, etc.

2. Open a WebSocket to the Qlik Associative Engine using the Dial function in the Dialer.

3. Open or create a document/app using openDoc or createApp.

Refer to the examples in https://github.com/qlik-oss/enigma-go/tree/master/examples for more information.

## Schemas

enigma-go includes generated API code that is based on the latest available Qlik Associative Engine schema.
When a new schema is available, a new version of enigma-go will be made available.
