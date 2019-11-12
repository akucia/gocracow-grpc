SERVER_FILES = server/server/*
PROTOS_ROOT = protos
SERVICE_PROTO = feature_store_service.proto
DS_META_PROTO = dataset_meta.proto
OUTPUT = bin/feature_store_server
PYTHON_ROOT = python
PYTHON_PROTO_DIR = featurestore/protos


go-client: proto client/go
		protoc -I=proto --go_out=plugins=grpc:client/go/photo_album proto/photo-album.proto
		go build -o photo-album-client-go client/go/photo-album-client.go

python-client: proto client/python
		python -m grpc_tools.protoc -I=proto --python_out=client/python/photo_album --grpc_python_out=client/python/photo_album proto/photo-album.proto
		ln -F -s ./client/python/photo-album-client.py photo-album-client-py

go-server: proto server
		protoc -I=proto --go_out=plugins=grpc:server/photo_album proto/photo-album.proto
		go build ./server/cmd/...

all: go-client python-client go-server