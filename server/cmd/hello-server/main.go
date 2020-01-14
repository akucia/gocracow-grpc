package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/akucia/gocracow-grpc/server/hello"
	"google.golang.org/grpc"
)

// HelloServer can greet the client
type HelloServer struct {
}

// Greetings says `Hello World!`
func (h HelloServer) Greetings(context.Context, *pb.Request) (*pb.Response, error) {
	log.Println("Received a call, greeting a client!")
	return &pb.Response{
		Text: "HelloWorld!",
	}, nil
}

var port = flag.Int("port", 5000, "Service port")

func main() {
	fmt.Println("Hello world - gRPC server")
	flag.Parse()
	// Open connection
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on :%d", *port)
	// Create new gRPC server
	grpcServer := grpc.NewServer()
	helloServer := HelloServer{}
	// Add our service to the service
	pb.RegisterHelloServer(grpcServer, &helloServer)
	// Start handling requests
	if err := grpcServer.Serve(lis); err != nil {
		panic("Unable to serve!")
	}
}
