package main

import (
	"log"
	"net"
	"rss-grpc/protos"

	"github.com/mmcdole/gofeed"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	parser := gofeed.NewParser()

	s := grpc.NewServer()
	rssServer := NewRssServer(parser)
	protos.RegisterRssServiceServer(s, rssServer)
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
