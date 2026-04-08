package rpcapi

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

var (
	GrpcClientConnection *grpc.ClientConn
)

func InitGrpcClient() {
	// Dial the gRPC server to get connection
	server := os.Getenv("GRPC_URL")
	if server == "" {
		server = "localhost:50051"
	}

	log.Printf("Connecting to gRPC server %s...\n", server)
	conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("Failed to connect to gRPC server: " + err.Error())
	}

	log.Printf("Successfully connected to server %s\n", server)
	GrpcClientConnection = conn
}

func CloseGrpcClient() {
	err := GrpcClientConnection.Close()
	if err != nil {
		log.Fatalln(err)
	}
}
