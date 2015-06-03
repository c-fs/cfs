package main

import (
	"github.com/c-fs/cfs/disk"
	pb "github.com/c-fs/cfs/proto"
	"github.com/qiniu/log"
	"golang.org/x/net/context"
)

type server struct {
	// server contains a map of disks.
	// The key in the map is the name of the disk.
	disks map[string]*disk.Disk
}

func NewServer() *server {
	return &server{disks: make(map[string]*disk.Disk)}
}

func (s *server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteReply, error) {
	dn, fn, err := splitDiskAndFile(req.Name)
	if err != nil {
		log.Infof("server: write error (%v)", err)
		return &pb.WriteReply{}, nil
	}

	d := s.Disk(dn)
	if d == nil {
		log.Infof("server: write error (cannot find disk %s)", dn)
		return &pb.WriteReply{}, nil
	}

	n, err := d.WriteAt(fn, req.Data, req.Offset)
	// TODO: add error
	if err != nil {
		log.Infof("server: write error (%v)", err)
		return &pb.WriteReply{}, nil
	}
	reply := &pb.WriteReply{BytesWritten: int64(n)}
	return reply, nil
}

func (s *server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadReply, error) {
	dn, fn, err := splitDiskAndFile(req.Name)
	if err != nil {
		log.Infof("server: read error (%v)", err)
		return &pb.ReadReply{}, nil
	}

	d := s.Disk(dn)
	if d == nil {
		log.Infof("server: read error (cannot find disk %s)", dn)
		return &pb.ReadReply{}, nil
	}

	// TODO: reuse buffer
	data := make([]byte, req.Length)
	n, err := d.ReadAt(fn, data, req.Offset)
	// TODO: add error
	if err != nil {
		log.Infof("server: read error (%v)", err)
		return &pb.ReadReply{}, nil
	}
	reply := &pb.ReadReply{BytesRead: int64(n), Data: data}
	return reply, nil
}

func (s *server) Rename(ctx context.Context, req *pb.RenameRequest) (*pb.RenameReply, error) {
	dn0, ofn, err := splitDiskAndFile(req.Oldname)
	if err != nil {
		log.Infof("server: rename error (%v)", err)
		return &pb.RenameReply{}, nil
	}
	dn1, nfn, err := splitDiskAndFile(req.Newname)
	if err != nil {
		log.Infof("server: rename error (%v)", err)
		return &pb.RenameReply{}, nil
	}
	if dn0 != dn1 {
		log.Infof("server: rename error (%v)", "not same disk")
		return &pb.RenameReply{}, nil
	}

	d := s.Disk(dn0)
	if d == nil {
		log.Infof("server: read error (cannot find disk %s)", dn0)
		return &pb.RenameReply{}, nil
	}

	err = d.Rename(ofn, nfn)
	if err != nil {
		log.Infof("server: rename error (%v)", err)
		return &pb.RenameReply{}, nil
	}
	reply := &pb.RenameReply{}
	return reply, nil
}

func (s *server) Remove(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveReply, error) {
	dn, fn, err := splitDiskAndFile(req.Name)
	if err != nil {
		log.Infof("server: remove error (%v)", err)
		return &pb.RemoveReply{}, nil
	}

	d := s.Disk(dn)
	if d == nil {
		log.Infof("server: remove error (cannot find disk %s)", dn)
		return &pb.RemoveReply{}, nil
	}

	err = d.Remove(fn, req.All)
	if err != nil {
		log.Infof("server: read error (%v)", err)
		return &pb.RemoveReply{}, nil
	}
	reply := &pb.RemoveReply{}
	return reply, nil
}

func (s *server) ReadDir(ctx context.Context, req *pb.ReadDirRequest) (*pb.ReadDirReply, error) {
	reply := &pb.ReadDirReply{}
	dn, fn, err := splitDiskAndFile(req.Name)
	if err != nil {
		log.Infof("server: readDir error (%v)", err)
		return reply, nil
	}

	d := s.Disk(dn)
	if d == nil {
		log.Infof("server: readDir error (cannot find disk %s)", dn)
		return reply, nil
	}

	stats, err := d.ReadDir(fn)
	if err != nil {
		log.Infof("server: readDir error (%v)", err)
		return reply, nil
	}

	reply.FileInfos = make([]*pb.FileInfo, len(stats))
	for i, stat := range stats {
		reply.FileInfos[i] = &pb.FileInfo{
			Name: stat.Name(),
			// TODO: Add size
			TotalSize: stat.Size(),
			IsDir:     stat.IsDir(),
		}
	}
	return reply, nil
}

func (s *server) Mkdir(ctx context.Context, req *pb.MkdirRequest) (*pb.MkdirReply, error) {
	reply := &pb.MkdirReply{}

	dn, fn, err := splitDiskAndFile(req.Name)
	if err != nil {
		log.Infof("server: mkdir error (%v)", err)
		return reply, nil
	}

	d := s.Disk(dn)
	if d == nil {
		log.Infof("server: mkdir error (cannot find disk %s)", dn)
		return reply, nil
	}
	err = d.Mkdir(fn, req.All)
	if err != nil {
		log.Infof("server: mkdir error (%v)", err)
		return reply, nil
	}
	return reply, nil
}
