SHELL := /bin/bash

test-unit:
	go test -count=1 -v -race .

lint:
	go fmt ./...

mod:
	go mod tidy

mod-update:
	go get -u ./...
	go mod tidy

generate: qix_generated.go

schema/engine-rpc.json:
	curl -s https://qlik.dev/specs/json-rpc/qix.json | jq > $@

qix_generated.go: schema/generate.go schema/engine-rpc.json
	go run ./schema/generate.go ./schema/engine-rpc.json ./schema/schema-companion.json $@ enigma disable-enigma-import
	go fmt $@
