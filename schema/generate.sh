#!/bin/bash

cd $(dirname "$0")

# Check that the EULA has been accepted
if [[ $ACCEPT_EULA != "yes" ]]; then
    echo "The EULA for Qlik Associative Engine was not accepted. Please check the README for instructions on how to accept the EULA"
    exit 1
fi

# Fetch latest published Qlik Associative Engine version on dockerhub if not specified
if [ -z $ENGINE_VERSION ]; then
    ENGINE_VERSION=$(curl -s "https://registry.hub.docker.com/v2/repositories/qlikcore/engine/tags/" | docker run -i stedolan/jq -r '."results"[0]["name"]' 2>/dev/null)
    echo "Using latest Engine version $ENGINE_VERSION"
fi

# Retrieve the JSON-RPC API from Qlik Associative Engine REST API
CONTAINER_ID=$(docker run -d -p 9077:9076 qlikcore/engine:$ENGINE_VERSION -S AcceptEULA=$ACCEPT_EULA)
RETRIES=0
while [[ $JSON_RPC_API == "" && $RETRIES != 10 ]]; do
    JSON_RPC_API=$(curl -fs localhost:9077/openapi/rpc)
    sleep 2
    RETRIES=$((RETRIES + 1 ))
done
docker kill $CONTAINER_ID

# Generate enigma-go based on schema
if [[ $JSON_RPC_API != "" ]]; then
    echo "Generating enigma-go based on JSON-RPC API for Qlik Associative Engine version $ENGINE_VERSION"
    echo "$JSON_RPC_API" > ./schema.json
    go run ./generate.go
    rm ./schema.json
    go fmt ../qix_generated.go > /dev/null
else
    echo "Failed to retrieve JSON-RPC API for Qlik Associative Engine version $ENGINE_VERSION"
    exit 1
fi

# To bump version, we could use something like this
# (which uses the bump.go file in bumper/ )
__version() {
  local ver
  local next_ver
  # relying on git describe requires us to always apply tags directly
  # to master
  ver=$(git describe 2> /dev/null)
  # 128 means that no tags were found, probably due to the fact that
  # there are none, so version is: 0.0.0.
  if [[ $? -eq 128 ]]; then
    ver="0.0.0"
  fi
  # Next version should be 0.1.0, bumped minor from 0.0.0
  next_ver=$(go run ./bumper/bump.go $ver -m)
  echo $next_ver
}
