#!/bin/bash

cd "$(dirname "$0")"

set -e

examples=(
    basics/custom-type/custom-type.go
    basics/lists/app-object-list/app-object-list.go
    basics/lists/field-list/field-list.go
    basics/lists/variable-list/variable-list.go
    basics/events/events.go
    cookiejar/cookiejar.go
    data/hypercubes/pivot/hypercube-pivot.go
    data/hypercubes/stacked/hypercube-stacked.go
    data/hypercubes/straight/hypercube-straight.go
    data/list-object/list-object.go
    data/string-expression/string-expression.go
    interceptors/retry-aborted/retry-aborted.go
    interceptors/metrics/metrics.go
    logging/traffic-log.go
    reload/monitor-progress/monitor-progress.go
    testing/mock-mode/mock-mode.go

)

for example in ${examples[@]}; do
	echo
	echo "Starting example: $example"
        go run $example
	echo "Ending example: $example"
done
