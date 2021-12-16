SHELL := /bin/bash

generate-schema:
	ACCEPT_EULA=yes ./schema/generate.sh

update-api-spec:
	go run ./spec/generate.go

update-engine-version: generate-schema update-api-spec
