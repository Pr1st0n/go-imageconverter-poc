FROM golang:1.12.8-alpine

# Install libvips
RUN apk add vips-dev fftw-dev build-base --update-cache \
    --repository https://alpine.global.ssl.fastly.net/alpine/edge/community/ \
    --repository https://alpine.global.ssl.fastly.net/alpine/edge/main

# Build application
WORKDIR /go/src/app

COPY ./imageconverter.go .

RUN apk add --no-cache git && \
    go get -u github.com/githubnemo/CompileDaemon && \
    go get -u github.com/davidbyttow/govips/pkg/vips

ENTRYPOINT CompileDaemon -log-prefix=false -build="go build -o imageconverter ." -command="./imageconverter"
