#!/bin/bash

APPNAME="wh"

APPVER="git describe --tags"


go build -o $PWD/$APPNAME -ldflags "-X main._version_=$(git describe --tags)" cmd/main.go

tar -cvf "$APPNAME.tar" $APPNAME script config/config.yml

rm -f $APPNAME