# Testing: MockMode

This example shows how to use MockMode and the TrafficDumpFile options.
To illustrate the concept, a scenario runs in two different modes.
1. __MockMode=false__ The first one, with the mock mode turned off, runs against a live QIX
Engine while generating a log file (specified by the TrafficDumpFile option) containing the protocol traffic.
2. __MockMode=true__ The second scenario runs with the mock mode turned on, which means that
it reads and replays the protocol traffic log files created in the first scenario

## Usage in test cases
MockMode is a concept that allows you to run integration tests in the ordinary go test suite without actually requiring a running Qlik Associative Engine. This is especially useful in cases where it might be difficult to set up a Qlik Associative Engine (or several Qlik Associative Engines for that matter) for the use case at hand.
Typically the MockMode parameter is read from an environment variable that specifies whether to run
against a live Qlik Associative Engine or not. Tests running in CI typically run against a live Qlik Associative Engine while the
default behaviour for a new developer should be replayed traffic since it is expected that  "go test ." should work out of the box.

