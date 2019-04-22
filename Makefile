# Binary name
BINARY=go-gecko
VERSION=G1-2.0

GITTAG=`git rev-parse --short HEAD`
BUILD_TIME=`date +%FT%T%z`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X gecko.Version=${VERSION}"
BUFLAGS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Release
BUILD_DIR=./build

# Binary
BUILD_OUT_ORIGIN="${BUILD_DIR}/${BINARY}.origin.bin"
BUILD_OUT_ZIP="${BUILD_DIR}/${BINARY}.bin"

# Builds the project
build:
		rm -rf ${BUILD_DIR}
		mkdir -p ${BUILD_DIR}

		# Build for linux
		${BUFLAGS} go build ${LDFLAGS} -o ${BUILD_OUT_ORIGIN} ./cmd/main.go

		# Compress
		upx -o ${BUILD_OUT_ZIP} ${BUILD_OUT_ORIGIN}
		rm ${BUILD_OUT_ORIGIN}

		# Copy configs and scripts
		cp -R ./cmd/conf.d ${BUILD_DIR}

		# Write version
		echo "${VERSION}" > ${BUILD_DIR}/version

install:
		go install

clean:
		go clean

.PHONY:  clean build