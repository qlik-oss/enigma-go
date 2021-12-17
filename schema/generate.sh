#!/bin/bash

cd $(dirname "$0")

if [[ -z "${QCS_HOST}" ]] || [[ -z "${QCS_API_KEY}" ]]; then
  echo "Please be sure to add a host (`QCS_HOST`) and apikey (`QCS_API_KEY`) as env. variable"
  exit 1
fi

OPEN_RPC_API=$(curl -fs -H "Authorization: Bearer $QCS_API_KEY" https://$QCS_HOST/api/engine/openrpc)

# Generate enigma-go based on schema
if [[ $OPEN_RPC_API != "" ]]; then
    echo "$OPEN_RPC_API" > ./schema.json
    ENGINE_VERSION=$(cat ./schema.json | jq -r '.info.version')
    echo "Generating enigma-go based on OpenRPC API for Qlik Associative Engine version $ENGINE_VERSION"
    go run ./generate.go ./schema.json  ./schema-companion.json ../qix_generated.go enigma disable-enigma-import
    rm ./schema.json
    go fmt ../qix_generated.go > /dev/null
else
    echo "Failed to retrieve OpenRPC API for Qlik Associative Engine version $ENGINE_VERSION"
    exit 1
fi
