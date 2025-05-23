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

generate:
	./schema/generate.sh
