# To bump version, we could use something like this
# (which uses the bump.go file in bumper/ )
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
  new_ver=$(go run ./verval/verval.go bump $ver)
  ecode=$?
  if [[ $ecode -ne 0 ]]; then
    echo $new_ver
    exit $ecode
  fi
  echo "New version: $new_ver"
  VERSION=v$new_ver+$(grep -oP "QIX_SCHEMA_VERSION.+?\K\d+\.\d+\.\d+" ../qix_generated.go)
}

__bump_version
echo $VERSION
echo "git tag -a ${VERSION} -m Release ${VERSION}"
git tag -a $VERSION -m "Release ${VERSION}"
