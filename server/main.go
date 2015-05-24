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

	cfs := NewServer()
	err = cfs.AddDisk("cfs0", "./cfs0000")
	if err != nil {
		log.Fatalf("server: failed to add disk (%v)", err)
	}

	s := grpc.NewServer()
	pb.RegisterCfsServer(s, cfs)
	s.Serve(lis)
}
