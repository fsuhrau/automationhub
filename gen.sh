#!/bin/sh
./protoc --proto_path=. --go_out=. --cpp_out=. action.proto
protoc --go_out=plugins=grpc:. *.proto