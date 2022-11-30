#!/bin/bash

set -e

build () {
  pushd ./src &>/dev/null 

#  go get honnef.co/go/tools/cmd/staticcheck

  gofmt -w .
#  staticcheck .
  CGO_ENABLED=1 GOARCH=amd64 go build -o ../lst_to_3d_linux_amd64 -v .

  popd &>/dev/null
}


$@
