#!/bin/bash
cd "$(dirname "$0")"

# Create coverage dir
rm -rf ./coverage
mkdir coverage

# Start the engine container
ENGINE_VERSION=$(curl -s "https://registry.hub.docker.com/v2/repositories/qlikcore/engine/tags/" | docker run -i stedolan/jq -r '."results"[0]["name"]' 2>/dev/null)
ENGINE_CONTAINER_ID=$(docker run -p9076:9076 -d qlikcore/engine:$ENGINE_VERSION -S AcceptEULA=yes)
docker cp ./examples/reload/monitor-progress/testdata/ $ENGINE_CONTAINER_ID:/testdata

# Execute both unit and integration tests
go test -coverprofile=./coverage/c.out -coverpkg="github.com/qlik-oss/enigma-go/v4" -race -p=1 ./... --count=1

# Stop the engine container
docker kill $ENGINE_CONTAINER_ID
docker rm $ENGINE_CONTAINER_ID

# Produce a coverage report
go tool cover -html=./coverage/c.out -o ./coverage/coverage.html

# Open in browser
open ./coverage/coverage.html
