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
	conn, err := grpc.Dial(fmt.Sprintf(":%d", *port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to gRPC service %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()
	client := pb.NewHelloClient(conn)
	for {
		response, err := client.Greetings(context.Background(), &pb.Request{})
		if err != nil {
			log.Fatalf("Unable to greet the server! %v", err)
		}
		log.Println(response.GetText())
		time.Sleep(time.Second * time.Duration(*repeat))
	}

}
