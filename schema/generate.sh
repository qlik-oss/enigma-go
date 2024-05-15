#!/bin/bash

## compares engine spec with master

generate() {
  if [[ $QCS_HOST == "" ]]; then
    echo "QCS_HOST not set"
    return 1
  fi
  if [[ $QCS_API_KEY  == "" ]]; then
    echo "QCS_API_KEY not set"
    return 1
  fi
  curl -s -H "Authorization: Bearer $QCS_API_KEY" https://$QCS_HOST/api/engine/openrpc | jq > ./schema/engine-rpc.json
  git diff --exit-code origin/master -- ./schema/engine-rpc.json >/dev/null
  if [[ $? -ne 0 ]]; then
    ENGINE_VERSION=$(cat ./schema/engine-rpc.json | jq -r '.info.version')
    echo "Generating enigma-go based on OPEN-RPC API for Qlik Associative Engine version $ENGINE_VERSION"
    ## generate code
    go run ./schema/generate.go ./schema/engine-rpc.json ./schema/schema-companion.json ./qix_generated.go enigma disable-enigma-import
    ## format code
    go fmt ./qix_generated.go >/dev/null
    ## generate spec
    go run ./spec/generate.go
  else
    echo "No changes to engine-rpc.json, nothing to do."
  fi
}

generate
