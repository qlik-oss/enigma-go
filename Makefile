SHELL := /bin/bash

generate-schema:
	./schema/generate.sh

update-api-spec:
	go run ./spec/generate.go

update-engine-version: generate-schema update-api-spec
