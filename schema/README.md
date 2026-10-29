# Schema Generator

This folder contains a program ([`generate.go`](./generate.go)) for generating enigma-go's `qix_generated.go` based on a specific version of Qlik Associative Engine.
The Engine JSON-RPC specification can be found in [`./engine-rpc.json`](./engine-rpc.json) and additional hand-written type information in [`./schema-companion.json`](./schema-companion.json).

To generate `qix_generated.go` from the current specification, run from the repository root:
```sh
make generate
```

To update to the latest version of the Engine specification (requires `curl` and `jq`):
```sh
rm schema/engine-rpc.json
make generate
```
This downloads the specification from https://qlik.dev/specs/json-rpc/qix.json before generating.

If you want to use a specific version of the specification instead, save it as `engine-rpc.json` in this folder and run `make generate`.
