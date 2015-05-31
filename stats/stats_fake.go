// +build !linux

package stats

import (
	pb "github.com/c-fs/cfs/proto"
	"golang.org/x/net/context"
)

func Server() pb.StatsServer { return &server{} }

type server struct{}

func (s *server) ContainerInfo(ctx context.Context, req *pb.ContainerInfoRequest) (*pb.ContainerInfoReply, error) {
	return &pb.ContainerInfoReply{Error: "unsupported to collect container info"}, nil
}
