#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"

project_dir="$(cd "$(dirname "${0}")/.." ; pwd)" # Absolute path to project dir

trap "rm -rf github.com" 0

proto_dirs=$(find ${project_dir}/proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep go_package $file &>/dev/null; then
      buf generate --template "${project_dir}/proto/buf.gen.gogo.yaml" $file
    fi
  done
done

# Remove old protobuf generated go files
find ${project_dir} -not -path "*github.com*" -and -name "*.pb*.go" -type f -delete

# move proto files to the right places
cp -r github.com/strangelove-ventures/async-icq/v5/* .
