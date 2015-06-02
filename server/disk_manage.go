package main

import (
	"os"
	"path"

	"github.com/c-fs/cfs/disk"
	"github.com/qiniu/log"
)

func (s *server) Disk(name string) *disk.Disk {
	return s.disks[name]
}

func (s *server) AddDisk(name, root string) error {
	s.disks[name] = &disk.Disk{Name: name, Root: root}
	err := os.MkdirAll(root, 0700)
	if err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Panicf("server: cannot get current working directory (%v)", err)
	}

	log.Infof("server: created disk[%s] at root path[%s]", name, path.Join(pwd, root))
	return nil
}
