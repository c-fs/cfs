package main

import (
	pb "github.com/c-fs/cfs/proto"
	"github.com/c-fs/cfs/server/config"
	"golang.org/x/net/context"
)

type metadataServer struct {
	disks []config.Disk
}

func (s *metadataServer) Disks(ctx context.Context, req *pb.DisksRequest) (*pb.DisksReply, error) {
	var disks []*pb.Disk
	for _, d := range s.disks {
		disks = append(disks, &pb.Disk{Name: d.Name})
	}
	return &pb.DisksReply{Disks: disks}, nil
}
