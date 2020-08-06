#!/bin/sh
go build .
./accountservice -profile=dev
