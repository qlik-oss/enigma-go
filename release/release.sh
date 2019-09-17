# !/bin/bash
# This script bumps the version based on the previous tag.
# If no tags are present, this will be interpreted '0.0.0'.
#
# After the version has been generated the tag will be added.
# Finally, a commit bumping the api spec to version=latest is added.
# Pushing the tag is left as an exercise to the reader.
#
# NOTE: This script should generate exactly 2 commits and 1 tag.

VERSION=""
bump_version() {
  local ver
  local new_ver
  # relying on git describe requires us to always apply tags directly
  # to master
  ver=$(git describe --tag --abbrev=0)
  ecode=$?
  # 128 means that no tags were found, probably due to the fact that
  # there are none, so version is: 0.0.0.
  if [[ $ecode -eq 128 ]]; then
    ver="0.0.0"
    echo "Tag not found, bumping from '$ver'"
  elif [[ $ecode -ne 0 ]]; then
    # we don't really know what's going on here...
    echo "Unexpected exit code from git describe!"
    return
  else
    echo "Found tag, bumping from '$ver'"
  fi
  new_ver=$(go run ./verval/verval.go bump $1 $ver)
  ecode=$?
  if [[ $ecode -ne 0 ]]; then
    echo $new_ver
    exit $ecode
  fi
  echo "New version: $new_ver"
  VERSION=v$new_ver
}

sanity_check() {
  if [[ ! -z $(git status --porcelain) ]]; then
    echo "There are uncommitted changes. Please make sure branch is clean."
    git status --porcelain
    exit 1
  fi
  local_branch=$(git rev-parse --abbrev-ref HEAD)
  if [[ $local_branch != "master" ]]; then
    echo "This script can only be run from the master branch."
    echo "You are on '$local_branch'. Aborting."
    exit 1
  fi
  # Check if local branch is up-to-date with remote master branch
  git fetch origin master
  git diff origin/master --exit-code > /dev/null
  if [[ $? -ne 0 ]]; then
    echo "Local branch is not up-to-date with remote master. Please pull the latest changes."
    git diff origin/master --name-only
    exit 1
  fi
}

if [[ $# -ne 1 ]]; then
  echo "use: release.sh <major|minor|patch>"
  exit 1
fi

case $1 in
  "major"|"minor"|"patch")
    sanity_check
    WD=$(pwd)
    cd $(dirname "$0")
    bump_version $1
    echo $VERSION
    cd ../spec
    echo -n "Generating spec..."
    go run generate.go -version=$VERSION
    if [[ $? -ne 0 ]]; then
      echo "FAIL"
      echo "Failed to generate API specification, aborting"
      exit 1
    fi
    echo "Done"
    QIX_VERSION=$(grep "QIX_SCHEMA_VERSION" qix_generated.go | cut -d ' ' -f4 | sed 's/"//g')
    MSG="Release: $VERSION for QIX schema version $QIX_VERSION"
    echo "git add ../api-spec.json"
    git add ../api-spec.json > /dev/null
    echo "git commit -m \"$MSG\""
    git commit -m "$MSG" > /dev/null
    echo "git tag -a ${VERSION} -m \"$MSG\""
    git tag -a $VERSION -m "$MSG" > /dev/null
    # Set version to latest on master
    echo "Bumping version of spec again, now to latest"
    go run generate.go
    echo "git add ../api-spec.json"
    git add ../api-spec.json > /dev/null
    echo "git commit -m \"Post-release: bumping version to latest\""
    git commit -m "Post-release: bumping version to latest" > /dev/null
    echo
    echo "If everything looks OK run the following command to release:"
    echo
    echo "  git push --follow-tags"
    echo
    cd $WD
    ;;
  *)
    echo "Argument must be one of 'major', 'minor' or 'patch'."
    exit 1
    ;;
esac
