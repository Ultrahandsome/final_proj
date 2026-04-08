package main

import (
	"CommentClassifier/internal/db"
	"CommentClassifier/internal/rpcapi"
	"CommentClassifier/internal/server"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	// Entrypoint of this application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init MongoDB
	err := db.InitMongo(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	// Init gRPC Client
	rpcapi.InitGrpcClient()

	// Init Gin web server
	server.InitServer()
	s := &http.Server{
		Addr:    ":38080",
		Handler: server.Server,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Graceful shutdown
	go func() {
		<-quit
		log.Println("Shutting down server...")
		if err := s.Close(); err != nil {
			log.Fatalln(err)
		}
		rpcapi.CloseGrpcClient()
	}()

	// Start server
	log.Println("Starting server...")
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}
}
