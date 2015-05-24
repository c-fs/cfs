package main

import (
	"log"

	"github.com/c-fs/cfs/disk"
	pb "github.com/c-fs/cfs/proto"
	"golang.org/x/net/context"
)

type server struct{}

func (s *server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteReply, error) {
	n, _ := disk.WriteAt(req.Name, req.Data, req.Offset)
	// TODO: add error
	reply := &pb.WriteReply{BytesWritten: int64(n)}
	return reply, nil
}

func (s *server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadReply, error) {
	// TODO: reuse buffer
	data := make([]byte, req.Length)
	n, err := disk.ReadAt(req.Name, data, req.Offset)
	// TODO: add error
	if err != nil {
		log.Printf("server: read error %v", err)
	}
	reply := &pb.ReadReply{BytesRead: int64(n), Data: data}
	return reply, nil
}
