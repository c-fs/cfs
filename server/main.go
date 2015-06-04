package main

import (
	"io/ioutil"
	"net"

	"github.com/BurntSushi/toml"
	pb "github.com/c-fs/cfs/proto"
	"github.com/c-fs/cfs/server/config"
	"github.com/c-fs/cfs/stats"
	"github.com/qiniu/log"
	"google.golang.org/grpc"
)

func main() {
	configFile := "default.conf"
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("server: cannot load configuration file[%s] (%v)", configFile, err)
	}

	var conf config.Server
	if _, err := toml.Decode(string(data), &conf); err != nil {
		log.Fatalf("server: configuration file[%s] is not valid (%v)", configFile, err)
	}

	log.Infof("server: starting server...")

	lis, err := net.Listen("tcp", net.JoinHostPort(conf.Bind, conf.Port))
	if err != nil {
		log.Fatalf("server: failed to listen: %v", err)
	}

	log.Infof("server: listening on %s", net.JoinHostPort(conf.Bind, conf.Port))

	cfs := NewServer()

	for _, d := range conf.Disks {
		err = cfs.AddDisk(d.Name, d.Root)
		if err != nil {
			log.Fatalf("server: failed to add disk (%v)", err)
		}
	}

	s := grpc.NewServer()
	pb.RegisterCfsServer(s, cfs)
	pb.RegisterStatsServer(s, stats.Server())
	log.Infof("server: ready to serve clients")
	s.Serve(lis)
}
