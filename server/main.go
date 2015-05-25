package main

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/BurntSushi/toml"
	pb "github.com/c-fs/cfs/proto"
	"github.com/c-fs/cfs/server/config"
	"google.golang.org/grpc"
)

func main() {
	data, err := ioutil.ReadFile("default.conf")
	if err != nil {
		log.Fatalf("server: cannot load configuration file[%s] (%v)", "default.conf", err)
	}

	var conf config.Server
	if _, err := toml.Decode(string(data), &conf); err != nil {
		log.Fatalf("server: configuration file[%s] is not valid (%v)", err)
	}

	log.Printf("server: starting server...")

	lis, err := net.Listen("tcp", net.JoinHostPort(conf.Bind, conf.Port))
	if err != nil {
		log.Fatalf("server: failed to listen: %v", err)
	}

	log.Printf("server: listening on %s", net.JoinHostPort(conf.Bind, conf.Port))

	cfs := NewServer()

	for _, d := range conf.Disks {
		err = cfs.AddDisk(d.Name, d.Root)
		if err != nil {
			log.Fatalf("server: failed to add disk (%v)", err)
		}
	}

	s := grpc.NewServer()
	pb.RegisterCfsServer(s, cfs)
	log.Printf("server: ready to serve clients")
	s.Serve(lis)
}
