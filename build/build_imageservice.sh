#!/bin/sh
echo "Build imageservice image"
# ROOT_DIR="$(dirname `pwd`)"
ROOT_DIR=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice

export GOOS=linux
export GOARCH=amd64
# Set to use static linking
export CGO_ENABLED=0

BUILD_DIR=${ROOT_DIR}/build/package/imageservice
# Go Build imageservice
cd ${ROOT_DIR}/cmd/imageservice
go get; go build
cp imageservice ${BUILD_DIR}/
echo "  go built `pwd`"

# Go Build healthchecker binary
cd ${ROOT_DIR}/cmd/healthchecker
go get; go build
cp healthchecker ${BUILD_DIR}/
echo "  go built `pwd`"

# Docker Build imageservice
cd ${ROOT_DIR}/build/package
docker build -t unusedprefix/imageservice imageservice/
