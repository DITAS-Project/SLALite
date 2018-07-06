#!/usr/bin/env sh
WORKDIR=$GOPATH/src/SLALite
mkdir $WORKDIR
cp -a . $WORKDIR
cd $WORKDIR
rm -rf vendor
dep ensure
CGO_ENABLED=0 GOOS=linux go build -a -o SLALite
go test ./...
cp SLALite $1