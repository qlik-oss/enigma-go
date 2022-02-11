SHELL := /bin/bash

test-unit:
	go test -count=1 -v -race .

generate-schema:
	./schema/generate.sh

update-api-spec:
	go run ./spec/generate.go

update-engine-version: generate-schema update-api-spec
