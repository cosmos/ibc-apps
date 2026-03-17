build:
    cargo build

test:
    cargo test

# builds the contracts, using the arm64 workspace-optimizer image if
# avaliable. as this switches images depending on the archetecture,
# this command should not be used to create release builds.
optimize:
    ./devtools/optimize.sh

simtest: optimize
    # normalize names for arm builds.
    if [[ $(uname -m) =~ "arm64" ]]; then \
    cp artifacts/polytone_note-aarch64.wasm artifacts/polytone_note.wasm && \
    cp artifacts/polytone_voice-aarch64.wasm artifacts/polytone_voice.wasm && \
    cp artifacts/polytone_tester-aarch64.wasm artifacts/polytone_tester.wasm && \
    cp artifacts/polytone_proxy-aarch64.wasm artifacts/polytone_proxy.wasm \
    ;fi

    mkdir -p tests/wasms
    cp -R ./artifacts/*.wasm tests/wasms

    go clean -testcache
    cd tests/simtests && go test ./...

integrationtest: optimize
    go clean -testcache
    cd tests/strangelove && go test ./...

# ${f    <-- from variable f
#   ##   <-- greedy front trim
#   *    <-- matches anything
#   /    <-- until the last '/'
#  }
# <https://stackoverflow.com/a/3162500>
schema:
    start=$(pwd); \
    for f in ./contracts/**/*; \
    do \
    echo "generating schema for ${f##*/}"; \
    cd "$f" && cargo schema && cd "$start" \
    ;done
