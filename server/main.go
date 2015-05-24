package main

import (
	"log"
	"net"

	pb "github.com/c-fs/cfs/proto"
	"google.golang.org/grpc"
)

const (
	// make this a flag
	port = ":15524"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("server: failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCfsServer(s, &server{})
	s.Serve(lis)
}
