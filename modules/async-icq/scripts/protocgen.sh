#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
cd proto

# generate gogo proto code
      buf generate --template buf.gen.gogo.yaml 

cd ..

# move proto files to the right places
cp -r github.com/cosmos/ibc-apps/modules/async-icq/v*/types/* types/
rm -rf github.com

# go mod tidy
