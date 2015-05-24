package main

import (
	"log"
	"os"

	"github.com/c-fs/cfs/disk"
	pb "github.com/c-fs/cfs/proto"
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

func (s *server) Disk(name string) *disk.Disk {
	return s.disks[name]
}

func (s *server) AddDisk(name, root string) error {
	s.disks[name] = &disk.Disk{Name: name, Root: root}
	err := os.MkdirAll(root, 0700)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteReply, error) {
	dn, fn, err := splitDiskAndFile(req.Name)
	if err != nil {
		log.Printf("server: write error (%v)", err)
		return &pb.WriteReply{}, nil
	}

	d := s.Disk(dn)
	if d == nil {
		log.Printf("server: write error (cannot find disk %s)", dn)
		return &pb.WriteReply{}, nil
	}

	n, _ := d.WriteAt(fn, req.Data, req.Offset)
	// TODO: add error
	reply := &pb.WriteReply{BytesWritten: int64(n)}
	return reply, nil
}

func (s *server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadReply, error) {
	dn, fn, err := splitDiskAndFile(req.Name)
	if err != nil {
		log.Printf("server: read error (%v)", err)
		return &pb.ReadReply{}, nil
	}

	d := s.Disk(dn)
	if d == nil {
		log.Printf("server: read error (cannot find disk %s)", dn)
		return &pb.ReadReply{}, nil
	}

	// TODO: reuse buffer
	data := make([]byte, req.Length)
	n, err := d.ReadAt(fn, data, req.Offset)
	// TODO: add error
	if err != nil {
		log.Printf("server: read error (%v)", err)
		return &pb.ReadReply{}, nil
	}
	reply := &pb.ReadReply{BytesRead: int64(n), Data: data}
	return reply, nil
}
