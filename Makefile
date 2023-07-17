SHELL := /bin/bash

test-unit:
	go test -count=1 -v -race .

generate-schema:
	./schema/generate.sh
