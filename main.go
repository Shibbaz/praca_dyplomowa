package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct{}

func NewServer() *server {
	return &server{}
}

func main() {
	// Create a listener on TCP port
	postsListen, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	serv := NewServer()
	// Create a gRPC server object
	postsServer := grpc.NewServer()

	// Attach the Greeter service to the server
	posts := RegisterPostsAPIServer(postsServer, serv)
	// Serve gRPC server
	log.Println("Serving posts/gRPC on 0.0.0.0:8081")

	go func() {
		log.Fatalln(postsServer.Serve(postsListen))
	}()

	// Create a client connection to the gRPC server we just started

	postsConn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:8081",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	// Create a new ServeMux for the gRPC-Gateway
	gwmux := runtime.NewServeMux()

	err = posts.RegisterPostsAPIHandler(context.Background(), gwmux, postsConn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	// Create a new HTTP server for the gRPC-Gateway
	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	log.Fatalln(gwServer.ListenAndServe())
}
