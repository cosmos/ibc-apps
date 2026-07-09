# make local-image
# docker run --rm -it rate-limiting:local q

FROM golang:1.26-alpine AS builder

RUN set -eux; apk add --no-cache git libusb-dev linux-headers gcc musl-dev make;

ENV GOPATH=""

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make build

FROM alpine:3.21

COPY --from=builder /go/build/* /bin/simd

ENTRYPOINT ["/bin/simd"]