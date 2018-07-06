#!/usr/bin/env sh
mkdir src
mkdir src/SLALite
cp -r *.* src/SLALite
export GOPATH=.
cd src/SLALite
rm -rf vendor
dep ensure
CGO_ENABLED=0 GOOS=linux go build -a -o SLALite
go test ./...
cp SLALite ../..