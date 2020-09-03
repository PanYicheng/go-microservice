#!/bin/sh
echo "Build accountservice image"
# ROOT_DIR="$(dirname `pwd`)"
ROOT_DIR=/home/pyc/gopath/src/github.com/PanYicheng/go-microservice

export GOOS=linux
export GOARCH=amd64
# Set to use static linking
export CGO_ENABLED=0

BUILD_DIR=${ROOT_DIR}/build/package/accountservice
# Go Build accountservice
cd ${ROOT_DIR}/cmd/accountservice
go get; go build
cp accountservice ${BUILD_DIR}/
echo "  go built `pwd`"

# Go Build healthchecker binary
cd ${ROOT_DIR}/cmd/healthchecker
go get; go build
cp healthchecker ${BUILD_DIR}/
echo "  go built `pwd`"

# Docker Build accountservice
cd ${ROOT_DIR}/build/package
docker build -t unusedprefix/accountservice accountservice/
