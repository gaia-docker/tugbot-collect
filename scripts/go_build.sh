#!/bin/bash
distdir=.dist

go_build() {
  rm -rf "${distdir}"
  mkdir "${distdir}"
  glide install
  CGO_ENABLED=0 go build -v -o ${distdir}/tugbot-collect
}

set -x
go_build
