#!/bin/sh

docker run --rm -v ${PWD}:/home -w /home znly/protoc --go_out=plugins=grpc:. -I. *.proto
