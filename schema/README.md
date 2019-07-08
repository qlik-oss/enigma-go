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

## Adding Allowed Arguments for "Enumtypes"

This section is for you who are developing and want to extend the set of allowed arguments for a property a given property.
Let's take an example working with a [HyperCubeDef](https://core.qlik.com/services/qix-engine/apis/qix/definitions/#hypercubedef). This definition has a property called `qMode` which is represented by an enum called `NxHypercubeMode`.

The allowed arguments will be restricted based on the different values that `NxHypercubeMode` can have. So if you want to extend the allowed arguments you can use:
```
AddArgumentsForType(NxHypercubeMode(""), []string{"DATA_MODE_FLUFFY"})
```
which will allow you to set `qMode: "DATA_MODE_FLUFFY"` when creating your `HyperCubeDef`.
Further information can be found in the comments in [enums.go](/enums.go) and examples in the related unittests: [enums_test.go](/enums_test.go).
