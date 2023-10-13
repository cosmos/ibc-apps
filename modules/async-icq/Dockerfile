# docker build . -t icqd:local
# docker run --rm -it icqd:local q

FROM golang:1.21-alpine3.18 as builder

RUN set -eux; apk add --no-cache git libusb-dev linux-headers gcc musl-dev make go;

ENV GOPATH=""

# ADD testing testing
# ADD LICENSE LICENSE

# COPY testing/contrib/devtools/Makefile contrib/devtools/Makefile
# COPY Makefile .

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go install -v ./...

FROM alpine:3.18

# echo go/bin/
RUN echo $PATH

COPY --from=builder /go/bin/* /bin/icqd

ENTRYPOINT ["icqd"]