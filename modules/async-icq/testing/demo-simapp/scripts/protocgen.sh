#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
cd proto

# generate gogo proto code
buf generate --template buf.gen.gogo.yaml 

cd ..

# move proto files to the right places
cp -r github.com/cosmos/ibc-apps/modules/async-icq/v7/interchain-query-demo/x/interquery/types/* x/interquery/types/
rm -rf github.com

go mod tidy
