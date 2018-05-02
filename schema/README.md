# Schema Generator

This folder contains a script for generating enigma-go based on a specific version of Qlik Analytics Engine.
The version to be used must be one of the published versions of the Qlik Analytics Engine, see [here](https://hub.docker.com/r/qlikcore/engine/tags/).

If a version is not specified the script will default to the latest published version.

```bash
ENGINE_VERSION=12.160.6 ./schema/generate.sh
```
