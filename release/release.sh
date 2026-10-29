# !/bin/bash
# This script bumps the version based on the previous tag.
# If no tags are present, this will be interpreted '0.0.0'.
# Use 'auto' to derive the bump from the conventional commits since the previous tag.
#
# After the version has been generated the tag will be added.
# Finally, a commit bumping the api spec to version=latest is added.
# Pushing the tag is left as an exercise to the reader.
#
# NOTE: This script should generate exactly 2 commits and 1 tag.

VERSION=""

# Determines the bump (major, minor or patch) from the conventional commit
# messages since the given tag (or from the whole history if the tag is empty):
#   breaking change ('!' after type/scope or a BREAKING CHANGE footer) -> major
#   feat                                                                -> minor
#   fix, perf                                                           -> patch
# Other types (chore, docs, ...) do not trigger a release on their own.
auto_bump() {
  local range=""
  [[ -n $1 ]] && range="$1..HEAD"
  local subjects bodies
  subjects=$(git log --format=%s $range)
  bodies=$(git log --format=%b $range)
  if grep -qE '^[a-zA-Z]+(\([^)]*\))?!:' <<< "$subjects" || grep -qE '^BREAKING[ -]CHANGE(:| )' <<< "$bodies"; then
    echo "major"
  elif grep -qE '^feat(\([^)]*\))?:' <<< "$subjects"; then
    echo "minor"
  elif grep -qE '^(fix|perf)(\([^)]*\))?:' <<< "$subjects"; then
    echo "patch"
  fi
}

bump_version() {
  local ver
  local new_ver
  local bump=$1
  local last_tag=""
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
    last_tag=$ver
  fi
  if [[ $bump == "auto" ]]; then
    bump=$(auto_bump "$last_tag")
    if [[ -z $bump ]]; then
      echo "No releasable commits (feat, fix, perf or breaking changes) found since '${last_tag:-the beginning}'."
      exit 1
    fi
    echo "Conventional commits since '${last_tag:-the beginning}' result in a '$bump' bump"
  fi
  new_ver=$(go run ./verval/verval.go bump $bump $ver)
  ecode=$?
  if [[ $ecode -ne 0 ]]; then
    echo $new_ver
    exit $ecode
  fi
  VERSION=v$new_ver
  echo "New version: $VERSION"
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
  echo "use: release.sh <major|minor|patch|auto>"
  exit 1
fi

case $1 in
  "major"|"minor"|"patch"|"auto")
    set -eo pipefail
    sanity_check
    WD=$(pwd)
    cd $(dirname "$0")
    bump_version $1
    QIX_VERSION=$(grep "QIX_SCHEMA_VERSION" ../qix_generated.go | cut -d ' ' -f4 | sed 's/"//g')
    if [[ -z $QIX_VERSION ]]; then
      echo "Couldn't find QIX schema version"
      exit 1
    fi
    MSG="Release: ${VERSION} for QIX schema version ${QIX_VERSION}"
    echo "git tag -a ${VERSION} -m \"${MSG}\""
    git tag -a ${VERSION} -m "${MSG}" > /dev/null
    # Set version to latest on master
    echo
    echo "If everything looks OK run the following command to release:"
    echo
    echo "  git push --follow-tags"
    echo
    cd $WD
    ;;
  *)
    echo "Argument must be one of 'major', 'minor', 'patch' or 'auto'."
    exit 1
    ;;
esac
