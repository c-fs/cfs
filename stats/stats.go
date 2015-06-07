package stats

import (
	"encoding/json"
	"sort"
	"strconv"

	pb "github.com/c-fs/cfs/proto"
	"github.com/codahale/metrics"
	"golang.org/x/net/context"
)

func init() {
	initContainerManager()
}

type CounterType struct {
	disk   string
	client string
	name   string
}

func Counter(name string) CounterType { return CounterType{name: name} }

func (c CounterType) Disk(disk string) CounterType {
	c.disk = disk
	return c
}

func (c CounterType) Client(client string) CounterType {
	c.client = client
	return c
}

func (c CounterType) Add() {
	var prefix string
	if c.disk != "" {
		prefix = c.disk + "_"
	}
	if c.client != "" {
		prefix = prefix + c.client + "_"
	}
	metrics.Counter(prefix + c.name).Add()
}

func Server() pb.StatsServer { return &server{} }

type server struct{}

func (s *server) ContainerInfo(ctx context.Context, req *pb.ContainerInfoRequest) (*pb.ContainerInfoReply, error) {
	info, err := containerInfo()
	if err != nil {
		return &pb.ContainerInfoReply{Error: err.Error()}, nil
	}
	b, err := json.Marshal(info)
	if err != nil {
		return &pb.ContainerInfoReply{Error: err.Error()}, nil
	}
	return &pb.ContainerInfoReply{Info: string(b)}, nil
}

func (s *server) Metrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsReply, error) {
	counters, _ := metrics.Snapshot()
	cms := make([]*pb.Metric, 0, len(counters))
	for n, v := range counters {
		cms = append(cms, &pb.Metric{Name: n, Val: strconv.FormatUint(v, 10)})
	}
	sort.Sort(metricsByName(cms))
	r := &pb.MetricsReply{Counters: cms}
	return r, nil
}

type metricsByName []*pb.Metric

func (p metricsByName) Len() int           { return len(p) }
func (p metricsByName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p metricsByName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
