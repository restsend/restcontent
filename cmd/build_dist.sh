#!/bin/sh

GIT_COMMIT=$(git rev-list -1 HEAD)
BUILD_TIME=$(date "+%Y-%m-%d_%H:%M:%S")

set -x
# build with gitcommit and build_time
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

go build -tags mysql\
   -ldflags "-X main.GitCommit=$GIT_COMMIT -X main.BuildTime=$BUILD_TIME" \
   -o restcontent .


mkdir app
mv restcontent app/
cp -r ../templates app/
cp -r ../static app/

tar -czvf restcontent-latest.tar.gz app
rm -Rf app
