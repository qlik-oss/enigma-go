# Schema Generator

This folder contains a script for generating enigma-go based on a specific version of Qlik Associative Engine.
The Engine OpenRPC specification can be found in the schema folder `./schema/engine-rpc.json`.

If you want to manually generate manually a specific version of the Engine specification:
- download the spec* and save it as `engine-rpc.json` in the `./schema` folder
- run:
```sh
./schema/generate.sh
```
> <sup>*</sup> You can download the Engine OpenRPC specification from `https://qlik.dev/apis/json-rpc/qix`
