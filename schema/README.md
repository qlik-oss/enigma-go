# Schema Generator

This folder contains a script for generating enigma-go based on a specific version of Qlik Associative Engine.
The version to be used must be one of the published versions of the Qlik Associative Engine, see [here](https://hub.docker.com/r/qlikcore/engine/tags/).

Please note that to be able to generate enigma-go you will need to accept the [EULA](https://core.qlik.com/eula/).

```bash
ACCEPT_EULA=<yes/no> ENGINE_VERSION=<version> ./schema/generate.sh
```

If a version is not specified the script will default to the latest published version.

```bash
ACCEPT_EULA=<yes/no> ./schema/generate.sh
```
