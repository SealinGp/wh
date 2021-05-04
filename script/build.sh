#!/bin/bash

APPNAME="wh"
APPVER=`git describe --tags`

go build -ldflags "-X 'main._version_=$APPVER'" -o $PWD/$APPNAME cmd/wh/main.go

tar -cvf "$APPNAME.tar" $APPNAME script config/wh.yml

rm -f $APPNAME