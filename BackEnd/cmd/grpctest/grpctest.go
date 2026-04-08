package main

import (
	"CommentClassifier/internal/rpcapi"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"math/rand"
	"net"
	"time"
)

// dummyServer implements the CommentClassifier gRPC service
type dummyServer struct {
	rpcapi.UnimplementedCommentClassifierServer
}

// ClassifyComments implements the gRPC method. For each RawComment received,
// it randomly picks a category, assigns a random confidence score, and generates three similar comments
// The response is streamed back as a ClassifiedComments message
func (s *dummyServer) ClassifyComments(req *rpcapi.RawComments, stream rpcapi.CommentClassifier_ClassifyCommentsServer) error {
	// Pre-defined labels to choose from.
	labels := []string{"Positive", "Negative", "Neutral", "Spam", "Needs Review"}

	// Process each incoming raw comment
	for _, rawComment := range req.Comments {
		// Randomly pick a label
		label := labels[rand.Intn(len(labels))]
		// Generate a random confidence score between 0.5 and 1.0
		score := 0.5 + rand.Float32()*0.5
		// Generate three similar comments
		similar := []string{
			"Similar comment A",
			"Similar comment B",
			"Similar comment C",
		}

		// Generate a keyword
		keywords := []string{
			"Keyword A",
			"Keyword B",
			"Keyword C",
		}

		// Create a ClassifiedComment for the raw comment
		classified := &rpcapi.ClassifiedComment{
			Id:             rawComment.Id,
			Label:          label,
			Score:          score,
			SimilarComment: similar,
			Keywords:       keywords,
		}

		// Wrap the classified comment in a ClassifiedComments message
		resp := &rpcapi.ClassifiedComments{
			Comments: []*rpcapi.ClassifiedComment{classified},
		}

		// Send the response back via the stream
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Allow a port to be set via a flag
	port := flag.String("port", "50051", "The server port")
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()
	// Register the dummy server as the CommentClassifier service
	rpcapi.RegisterCommentClassifierServer(grpcServer, &dummyServer{})

	// Enable reflection for debugging (e.g., using grpcurl)
	reflection.Register(grpcServer)

	log.Printf("Dummy gRPC server listening on port %s", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
