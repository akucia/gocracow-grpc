package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/akucia/gocracow-grpc/client/go/hello"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 5000, "Service port")
var repeat = flag.Int("n", 5, "Repeat call every n seconds")

func main() {
	fmt.Println("Hello world - go client")
	flag.Parse()
	// Create a new connection to the server
	conn, err := grpc.Dial(fmt.Sprintf(":%d", *port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to gRPC service %v", err)
	}
	// let's be nice and close the connection in the end
	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()

	// New client of Hello service
	client := pb.NewHelloClient(conn)
	for {
		// Call the method
		response, err := client.Greetings(context.Background(), &pb.Request{})
		if err != nil {
			log.Fatalf("Unable to greet the server! %v", err)
		}
		// Read the response fields
		log.Println(response.GetText())
		time.Sleep(time.Second * time.Duration(*repeat))
	}

}
