package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"

	pb "github.com/akucia/gocracow-grpc/client/go/photos"
	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/lucasb-eyer/go-colorful"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

func main() {
	app := &cli.App{
		Name:  "gRPC Photos Service client",
		Usage: "service API definition can be found in photos.proto",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Value: 5000,
				Usage: "service port",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list photos stored by the service",
				Action: func(c *cli.Context) error {
					client, closingFunc := newClient(c.Int("port"))
					defer closingFunc()
					response, err := client.ListPhotos(context.Background(), &pb.ListPhotosRequest{})
					if err != nil {
						log.Fatal(err)
					}
					for _, photoID := range response.GetPhotoIds() {
						fmt.Println(photoID)
					}
					return nil
				},
			},
			{
				Name:    "download",
				Aliases: []string{"d"},
				Usage:   "download photo and display in terminal",
				Action: func(c *cli.Context) error {
					client, closingFunc := newClient(c.Int("port"))
					defer closingFunc()
					var photoIDs []int32
					for _, photoID := range c.Args().Slice() {
						i, err := strconv.Atoi(photoID)
						if err != nil {
							return err
						}
						photoIDs = append(photoIDs, int32(i))
					}
					request := pb.GetAllPhotosRequest{PhotoIds: photoIDs}
					stream, err := client.GetAllPhotos(context.Background(), &request)
					if err != nil {
						return err
					}
					for {
						photoResponse, err := stream.Recv()
						if err == io.EOF {
							log.Info("end of photo stream")
							break
						}
						if err != nil {
							return err
						}
						err = showPhoto(*photoResponse.GetPhoto())
						if err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:    "upload",
				Aliases: []string{"u"},
				Usage:   "upload photo to the server",
				Action: func(c *cli.Context) error {
					// parse arguments from terminal
					photoPaths := c.Args().Slice()
					// new connection
					client, closingFunc := newClient(c.Int("port"))
					defer closingFunc()
					// open stream
					stream, err := client.PutPhotos(context.Background())
					if err != nil {
						return err
					}

					waitChan := make(chan struct{})

					// Receiving
					go func() {
						for {
							response, err := stream.Recv()
							if err == io.EOF {
								// read done.
								close(waitChan)
								return
							}
							// communication error
							if err != nil {
								log.Fatalf("Failed to receive a response : %v", err)
							}
							// communication was ok, but the file was corrupted
							if response.GetStatus() != pb.Status_OK {
								log.Printf(
									"server didn't confirm the photo, perhaps we should retry sending %v\n",
									response.GetFilename(),
								)
							} else {
								// All good!
								log.Printf("%v OK\n", response.GetFilename())
							}
						}
					}()
					// Sending
					for _, photoPath := range photoPaths {
						log.Printf("sending %v\n", photoPath)
						dat, err := ioutil.ReadFile(photoPath)
						if err != nil {
							return fmt.Errorf("unable to read file %v: %v", photoPath, err)
						}
						// create request
						photo := &pb.PutPhotoRequest{
							Filename: photoPath,
							Content:  dat,
						}
						// send the request using the stream
						if err := stream.Send(photo); err != nil {
							return fmt.Errorf("failed to send a photo: %v", err)
						}

					}
					// cleanup
					if err := stream.CloseSend(); err != nil {
						return err
					}
					<-waitChan
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// newClient creates a new Photos service client on localhost and specified port
func newClient(port int) (pb.PhotosClient, func()) {
	conn, err := grpc.Dial(fmt.Sprintf(":%d", port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to gRPC service %v", err)
	}
	closingFunc := func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}
	client := pb.NewPhotosClient(conn)
	return client, closingFunc
}

// showPhoto draws photo in terminal
func showPhoto(photo pb.Photo) error {
	dm := ansimage.NoDithering
	reader := bytes.NewReader(photo.GetContent())
	bg, err := colorful.Hex("#000000")
	if err != nil {
		return err
	}
	tx, ty := 15, 10
	sfy, sfx := ansimage.BlockSizeY, ansimage.BlockSizeX
	sm := ansimage.ScaleMode(2)
	pix, err := ansimage.NewScaledFromReader(reader, sfy*ty, sfx*tx, bg, sm, dm)
	if err != nil {
		return err
	}
	pix.SetMaxProcs(runtime.NumCPU()) // maximum number of parallel goroutines!
	pix.DrawExt(false, false)
	return nil
}
