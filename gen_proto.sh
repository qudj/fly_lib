#!/bin/bash
protoc --gogofaster_out=plugins=grpc:. --proto_path=proto fcc.proto
mkdir -p models/proto/fcc_serv
mv fcc.pb.go models/proto/fcc_serv/

protoc --gogofaster_out=plugins=grpc:. --proto_path=proto fly_starling.proto
mkdir -p models/proto/fly_starling_serv
mv fly_starling.pb.go models/proto/fly_starling_serv/