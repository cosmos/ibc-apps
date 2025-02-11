# docker build . -t pfm:local
# docker run --rm -it pfm:local q ibc-router

<<<<<<< HEAD
FROM golang:1.22-alpine3.20 as builder
=======
FROM golang:1.21-alpine3.18 AS builder
>>>>>>> 515bdca (fix: don't unescrow tokens that are moved between escrow accounts (#230))

RUN set -eux; apk add --no-cache git libusb-dev linux-headers gcc musl-dev make;

ENV GOPATH=""

ADD testing testing
ADD LICENSE LICENSE

COPY testing/contrib/devtools/Makefile contrib/devtools/Makefile
COPY Makefile .

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make build

FROM alpine:3.18

COPY --from=builder /go/build/simd /bin/simd

ENTRYPOINT ["simd"]

