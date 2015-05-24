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
	log.Printf("server: starting server...")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("server: failed to listen: %v", err)
	}

	log.Printf("server: listening on %s", port)

	cfs := NewServer()
	err = cfs.AddDisk("cfs0", "./cfs0000")
	if err != nil {
		log.Fatalf("server: failed to add disk (%v)", err)
	}

	s := grpc.NewServer()
	pb.RegisterCfsServer(s, cfs)
	log.Printf("server: ready to serve clients")
	s.Serve(lis)
}
