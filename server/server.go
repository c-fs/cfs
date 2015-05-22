package server

import (
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
