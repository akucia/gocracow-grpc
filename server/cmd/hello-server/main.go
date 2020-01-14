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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on :%d", *port)
	grpcServer := grpc.NewServer()
	helloServer := HelloServer{}
	pb.RegisterHelloServer(grpcServer, &helloServer)
	if err := grpcServer.Serve(lis); err != nil {
		panic("Unable to serve!")
	}
}
