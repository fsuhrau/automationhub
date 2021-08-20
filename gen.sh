#!/bin/sh
SRC_DIR=./proto
DST_DIR=./out
mkdir -p $DST_DIR/csharp $DST_DIR/cpp

protoc -I=$SRC_DIR --go_out=. --cpp_out=$DST_DIR/cpp --csharp_out=$DST_DIR/csharp $SRC_DIR/action.proto
# protoc --go_out=plugins=grpc:. *.proto