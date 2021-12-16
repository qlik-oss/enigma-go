#!/bin/bash

## compares engine spec with master
git diff --exit-code origin/master -- ./spec/engine-rpc.json > /dev/null
if [[ $? -ne 0 ]]; then
  ENGINE_VERSION=$(cat ./spec/engine-rpc.json | jq -r '.info.version')
  MSG "Generating enigma-go based on OPEN-RPC API for Qlik Associative Engine version $ENGINE_VERSION"
  echo $MSG
  go run ./schema/generate.go ./spec/engine-rpc.json ./schema/schema-companion.json ./qix_generated.go enigma disable-enigma-import
  go fmt ./qix_generated.go > /dev/null

  git add . --dry-run
  git commit -a -m "chore: ${MSG}" --dry-run
fi
