#!/bin/bash

export addr=":1234"

./wh -c "config/wh.yml" -s "$addr" -e "prod"