#!/bin/bash

set -o errexit          # Exit on most errors (see the manual)
set -o errtrace         # Make sure any error trap is inherited
set -o nounset          # Disallow expansion of unset variables
set -o pipefail         # Use last non-zero exit code in a pipeline
#set -o xtrace          # Trace the execution of the script (debug)

# Change directory to the project root
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR/.."
pwd

goa gen github.com/keboola/temp-webhooks-api/api/webhooks --output ./internal/pkg/webhooks/api
	rm -rf ./internal/pkg/webhooks/api/gen/http/cli
	rm -rf ./internal/pkg/webhooks/api/gen/http/webhooks/client
	mv ./internal/pkg/webhooks/api/gen/http/openapi* ./api/webhooks/gen
