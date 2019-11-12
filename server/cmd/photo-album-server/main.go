package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/akucia/gocracow-grpc/server/photo_album"
	"google.golang.org/grpc"
)

// Album holds photos from the trip
type Album struct {
}

// GetPhoto returns a photo from Album
func (a *Album) GetPhoto(context.Context, *photo_album.GetPhotoRequest) (*photo_album.GetPhotoResponse, error) {
	fmt.Println("GetPhoto")
	return nil, nil
}

// ListPhotos returns all photos in Album
func (a *Album) ListPhotos(context.Context, *photo_album.ListPhotosRequest) (*photo_album.ListPhotosResponse, error) {
	fmt.Println("ListPhotos")
	return nil, nil
}

func main() {
	fmt.Println("Hello world photo server!")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8888))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	// albumService := Album{}
	photo_album.RegisterAlbumServiceServer(grpcServer, &Album{})
	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
