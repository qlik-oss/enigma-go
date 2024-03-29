#!/bin/bash

## compares engine spec with master

wget -q https://qlik.dev/specs/json-rpc/qix.json -O ./schema/engine-rpc.json
git diff --exit-code origin/master -- ./schema/engine-rpc.json >/dev/null
if [[ $? -ne 0 ]]; then
    ENGINE_VERSION=$(cat ./schema/engine-rpc.json | jq -r '.info.version')
    MSG="Generating enigma-go based on OPEN-RPC API for Qlik Associative Engine version $ENGINE_VERSION"
    echo $MSG
    ## generate code
    go run ./schema/generate.go ./schema/engine-rpc.json ./schema/schema-companion.json ./qix_generated.go enigma disable-enigma-import
    ## format code
    go fmt ./qix_generated.go >/dev/null
    ## generate spec
    go run ./spec/generate.go
    ## configure git
    if [ "$CIRCLECI" == true ]; then
        git config --global user.email "no-reply@example.com"
        git config --global user.name "github-actions-bot"
        git add .
        git commit -a -m "chore: ${MSG}"
        git push --set-upstream origin ${CIRCLE_BRANCH}
    fi
else
    echo "No changes to engine-rpc.json, nothing to do."
fi
