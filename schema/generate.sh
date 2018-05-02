#!/bin/bash

cd $(dirname "$0")

# Fetch latest published Qlik Analytics Engine version on dockerhub if not specified
if [ -z $ENGINE_VERSION ]; then
    ENGINE_VERSION=$(curl -s "https://registry.hub.docker.com/v2/repositories/qlikcore/engine/tags/" | docker run -i stedolan/jq -r '."results"[0]["name"]' 2>/dev/null)
    echo "Using latest Engine version is $ENGINE_VERSION"
fi

# Retrieve the JSON-RPC API from Qlik Analytics Engines REST API
CONTAINER_ID=$(docker run -d -p 9076:9076 qlikcore/engine:$ENGINE_VERSION -S AcceptEULA=yes)
RETRIES=0
while [[ $JSON_RPC_API == "" && $RETRIES != 10 ]]; do
  JSON_RPC_API=$(curl -fs localhost:9076/jsonrpc-api)
  sleep 2
  RETRIES=$((RETRIES + 1 ))
done
docker kill $CONTAINER_ID

# Generate enigma-go based on schema
if [[ $JSON_RPC_API != "" ]]; then
    echo "Generating enigma-go based on JSON-RPC API for Qlik Analytics Engine version $ENGINE_VERSION"
    echo "$JSON_RPC_API" > ./schema.json
    go run ./generate.go
    rm ./schema.json
else
    echo "Failed to retrieve JSON-RPC API for Qlik Analytics Engine version $ENGINE_VERSION"
    exit 1
fi
