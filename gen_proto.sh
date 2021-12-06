#!/bin/bash
protoc --gogofaster_out=plugins=grpc:. --proto_path=/Users/qudongjie/go/src/github.com/qudj/fly_proto fcc.proto
mkdir -p models/proto/fcc_serv
mv fcc.pb.go models/proto/fcc_serv/