SERVER_FILES = server/server/*
PROTOS_ROOT = protos
SERVICE_PROTO = feature_store_service.proto
DS_META_PROTO = dataset_meta.proto
OUTPUT = bin/feature_store_server


hello-client: proto client/go/hello client/go/hello-client.go
		mkdir -p client/go/hello
		protoc -I=proto --go_out=plugins=grpc:client/go/hello proto/hello.proto
		go build -o hello-client ./client/go/hello-client.go

hello-server: proto server/cmd/hello-server/
		mkdir -p server/hello
		protoc -I=proto --go_out=plugins=grpc:server/hello proto/hello.proto
		go build ./server/cmd/hello-server/

photo-client: proto client/go
		mkdir -p client/go/photos
		protoc -I=proto --go_out=plugins=grpc:client/go/photos proto/photos.proto
		go build -o photos-client ./client/go/photos-client.go

photo-server: proto server
		mkdir -p server/server
		protoc -I=proto --go_out=plugins=grpc:server/photos proto/photos.proto
		go build ./server/cmd/photos-server/

hello: hello-client hello-server

photos: photo-client photo-server

all: hello photos