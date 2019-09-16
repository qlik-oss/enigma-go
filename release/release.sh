# !/bin/bash
# This script bumps the version based on the previous tag
# and appends engine version as metadata. If no tags are
# present, this will be interpreted '0.0.0'.

# After the version has been generated the tag will be added.
# Pushing the tag is left as an exercise to the reader.

VERSION=""
__bump_version() {
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
  VERSION=v$new_ver+$(grep -oP "QIX_SCHEMA_VERSION.+?\K\d+\.\d+\.\d+" ../qix_generated.go)
}

if [[ $# -ne 1 ]]; then
  echo "use: release.sh <major|minor|patch>"
  exit 1
fi

# TODO check sanity!

case $1 in
  "major"|"minor"|"patch")
    __bump_version $1
    echo $VERSION
    WD=$(pwd)
    cd ../spec
    echo -n "Generating spec..."
    go run generate.go -version=$VERSION
    if [[ $? -ne 0 ]]; then
      echo "FAIL"
      echo "Failed to generate API specification, aborting"
      exit 1
    fi
    echo "Done"
    echo "git add ../api-spec.json"
    git add ../api-spec.json > /dev/null
    echo "git commit -m \"Release: $VERSION\""
    git commit -m "Release: $VERSION" > /dev/null
    echo "git tag -a ${VERSION} -m Release ${VERSION}"
    git tag -a $VERSION -m "Release ${VERSION}" > /dev/null
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
