# host
# make local-image
# docker run --rm -it ibchooks:local q

FROM golang:1.21-alpine3.18 as builder

RUN set -eux; apk add --no-cache ca-certificates git build-base libusb-dev linux-headers gcc musl-dev make go;

ENV GOPATH=""

COPY go.mod .
COPY go.sum .

RUN set -eux; \    
    export ARCH=$(uname -m); \
    WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
    if [ ! -z "${WASM_VERSION}" ]; then \
      wget -O /lib/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.${ARCH}.a; \      
    fi; \
    go mod download;


COPY . .

RUN LINK_STATICALLY=true BUILD_TAGS=muslc make build

FROM alpine:3.18

COPY --from=builder /go/build/simd /bin/simd

ENTRYPOINT ["/bin/simd"]