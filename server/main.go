package main

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/c-fs/cfs/enforce"
	pb "github.com/c-fs/cfs/proto"
	"github.com/c-fs/cfs/server/config"
	"github.com/c-fs/cfs/stats"
	"github.com/qiniu/log"
	"google.golang.org/grpc"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	configfn := flag.String("config", "default.conf", "location of configuration file")
	flag.Parse()

	data, err := ioutil.ReadFile(*configfn)
	if err != nil {
		log.Fatalf("server: cannot load configuration file[%s] (%v)", *configfn, err)
	}

	var conf config.Server
	if _, err := toml.Decode(string(data), &conf); err != nil {
		log.Fatalf("server: configuration file[%s] is not valid (%v)", *configfn, err)
	}

	// default is that cfs is bootstrapped using docker
	cname, err := detectDockerContainer()
	if err != nil {
		log.Printf("server: failed to detect docker container (%v)", err)
	} else {
		stats.SetContainerName(cname)
		log.Printf("server: detect docker container %q", cname)
	}

	log.Infof("server: starting server...")

	lis, err := net.Listen("tcp", net.JoinHostPort(conf.Bind, conf.Port))
	if err != nil {
		log.Fatalf("server: failed to listen: %v", err)
	}

	log.Infof("server: listening on %s", net.JoinHostPort(conf.Bind, conf.Port))

	cfs := NewServer()

	for _, d := range conf.Disks {
		name := d.Name
		if name == "" {
			name = strconv.FormatInt(rand.Int63(), 16)
		}
		err = cfs.AddDisk(name, d.Root)
		if err != nil {
			log.Fatalf("server: failed to add disk (%v)", err)
		}
	}

	// 0x1234 is the client ID for cfsctl, and its quota is 10 req/sec.
	enforce.SetQuota(0x1234, 10)

	// TODO report with influxSinker
	stats.Report(nil, 3*time.Second)

	s := grpc.NewServer()
	pb.RegisterCfsServer(s, cfs)
	pb.RegisterMetadataServer(s, &metadataServer{disks: conf.Disks})
	pb.RegisterStatsServer(s, stats.Server())
	log.Infof("server: ready to serve clients")
	s.Serve(lis)
}
