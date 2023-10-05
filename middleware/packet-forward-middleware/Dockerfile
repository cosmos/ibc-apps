# docker build . -t pfm:local
# docker run --rm -it pfm:local q ibc-router

FROM golang:1.21-alpine3.18 as builder

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

