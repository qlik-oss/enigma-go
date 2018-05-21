#!/bin/bash

cd $(dirname "$0")

repo_folder=enigma-go-repo
branch_name=automated-bump-engine-schema

if [ -z "$GH_TOKEN" ]; then
  echo "Missing GH_TOKEN variable."
  exit 1
fi

rm -rf $repo_folder
mkdir $repo_folder
cd $repo_folder

git init
git remote add origin git@github.com:qlik-oss/enigma-go.git
git config core.fileMode false
git config core.autocrlf false
git config core.safecrlf false
git fetch
git checkout master

existing_branch=$(git branch -r | grep -i "$branch_name")

if [ ! -z "$existing_branch" ]; then
  echo "There is already a bump branch, please remove/merge it and run this tool again."
  exit 1
fi

# Generate enigma-go based on latest published Qlik Associative Engine image
. ./schema/generate.sh

# If there are changes to qix_generated.go then open a pull request
local_changes=$(git ls-files qix_generated.go -m)

if [ ! -z "$local_changes" ]; then
  git checkout -b $branch_name
  git add qix_generated.go
  git commit -m "Automated: New API based on $ENGINE_VERSION"
  git push -u origin $branch_name
  curl -u none:$GH_TOKEN https://api.github.com/repos/qlik-oss/enigma-go/pulls --request POST --data "{
        \"title\": \"Automated: Generated enigma-go based on new JSON-RPC API\",
        \"body\": \"Hello! This is an automated pull request.\n\nI have generated a new enigma-go based on the JSON-RPC API for Qlik Associative Engine version $ENGINE_VERSION.\",
        \"head\": \"$branch_name\",
        \"base\": \"master\"
      }"
else
  echo "No changes to schema."
fi
