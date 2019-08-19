FROM golang:1.12.8-alpine

# Build libvips
WORKDIR /usr/local/src

RUN apk update && \
	apk upgrade

RUN apk add build-base \
	autoconf \
	automake \
	libtool \
	bc \
	zlib-dev \
	libxml2-dev \
	jpeg-dev \
	openjpeg-dev \
	tiff-dev \
	glib-dev \
	gdk-pixbuf-dev \
	sqlite-dev \
	libjpeg-turbo-dev \
	libexif-dev \
	lcms2-dev \
	fftw-dev \
	giflib-dev \
	libpng-dev \
	libwebp-dev \
	orc-dev \
	poppler-dev \
	librsvg-dev \
	libgsf-dev \
	openexr-dev \
	gtk-doc

ARG VIPS_VERSION=8.8.1
ARG VIPS_URL=https://github.com/libvips/libvips/releases/download

RUN wget ${VIPS_URL}/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz && \
	tar xf vips-${VIPS_VERSION}.tar.gz && \
	cd vips-${VIPS_VERSION} && \
	./configure && \
	make V=0 && \
	make install

# Build application
WORKDIR /go/src/app

COPY ./imageconverter.go .

RUN apk add --no-cache git && \
    go get github.com/derekparker/delve/cmd/dlv && \
    go get -u github.com/githubnemo/CompileDaemon && \
    go get -u github.com/davidbyttow/govips/pkg/vips

ENTRYPOINT CompileDaemon -log-prefix=false -build="go build -o imageconverter ." -command="./imageconverter"
