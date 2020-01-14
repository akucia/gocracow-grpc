package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"net"
	"sync"

	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"

	_ "image/jpeg"
	_ "image/png"

	pb "github.com/akucia/gocracow-grpc/server/photos"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// PhotoService holds all our gopher photos
type PhotoService struct {
	photos []*pb.Photo
	mux    sync.Mutex
}

// ListPhotos responds with the list of all photo IDs
func (p *PhotoService) ListPhotos(context.Context, *pb.ListPhotosRequest) (*pb.ListPhotosResponse, error) {
	var IDs []int32
	for _, photo := range p.photos {
		IDs = append(IDs, photo.GetPhotoId())
	}
	response := pb.ListPhotosResponse{PhotoIds: IDs}
	return &response, nil
}

// GetAllPhotos streams requested photos
func (p *PhotoService) GetAllPhotos(request *pb.GetAllPhotosRequest, stream pb.Photos_GetAllPhotosServer) error {
	for _, photo := range p.photos {
		for _, photoID := range request.GetPhotoIds() {
			if photo.GetPhotoId() == photoID {
				log.Infof("sending photo %v\n", photoID)
				response := &pb.GetPhotoResponse{Photo: photo}
				if err := stream.Send(response); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// PutPhotos reads photos from stream and saves then in the service.
// A response with status for each photo is sent back to the client to
// let it know if operation was successful.
func (p *PhotoService) PutPhotos(stream pb.Photos_PutPhotosServer) error {
	for {
		request, err := stream.Recv()
		// stream is closed
		if err == io.EOF {
			return nil
		}
		// communication error
		if err != nil {
			return err
		}
		log.Printf("received %v\n", request.GetFilename())
		var status pb.Status
		reader := bytes.NewReader(request.GetContent())
		// our photo validation
		img, _, err := image.Decode(reader)
		if err != nil {
			log.Warn(err)
			// corrupted image file
			status = pb.Status_ERROR
		} else {
			// store the image
			p.mux.Lock()
			newPhoto := &pb.Photo{
				PhotoId:   int32(len(p.photos)),
				Filename:  request.GetFilename(),
				Content:   request.GetContent(),
				Height:    int32(img.Bounds().Max.Y),
				Width:     int32(img.Bounds().Max.X),
				Timestamp: ptypes.TimestampNow(),
			}
			p.photos = append(p.photos, newPhoto)
			status = pb.Status_OK
			p.mux.Unlock()
		}
		// create response
		response := &pb.PutPhotoResponse{
			Filename: request.GetFilename(),
			Status:   status,
		}
		// send it back to the client
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func main() {
	log.Info("photos server starting...")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 5000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// additional middleware
	// makes the gRPC logs visible
	logrusEntry := logrus.NewEntry(logrus.New())
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)
	opts := []grpc.ServerOption{
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_logrus.StreamServerInterceptor(logrusEntry),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_logrus.UnaryServerInterceptor(logrusEntry),
			),
		),
	}
	grpcServer := grpc.NewServer(opts...)
	photoService := PhotoService{}
	pb.RegisterPhotosServer(grpcServer, &photoService)
	log.Info("photos server is waiting for requests...")
	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}

}
