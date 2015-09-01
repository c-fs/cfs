package stats

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	pb "github.com/c-fs/cfs/proto"
	"github.com/codahale/metrics"
	"golang.org/x/net/context"
)

// The recommended way to run cfs is putting the process in an
// exclusive container, which has its own cgroup in each hierarchy.
// It helps cfs to monitor its resource usage.
//
// It asks users to specify container name instead of detecting
// cgroup that has the process automatically because it may meet
// strange cases:
// 1. cfs process may be in different cpu/memory/etc cgroups
// 2. the cgroup that includes cfs may have other processes
// So it hopes that user could take care of it.
const (
	DefaultContainerName = "/cfs"
	clientCounterPrefix  = "client_"
)

var containerName string

func init() {
	initContainerManager()
	containerName = DefaultContainerName
}

func SetContainerName(name string) { containerName = name }

type CounterType struct {
	disk   string
	client int64
	op     string
}

func Counter(disk, op string) *CounterType {
	return &CounterType{
		op:   op,
		disk: disk,
	}
}

func (c *CounterType) Client(id int64) *CounterType {
	c.client = id
	return c
}

func (c *CounterType) Add() {
	metrics.Counter(c.disk + "_" + c.op).Add()
	if c.client != 0 {
		metrics.Counter(ClientCounterName(c.client)).Add()
	}
}

func ClientCounterName(id int64) string {
	return clientCounterPrefix + strconv.FormatInt(id, 16)
}

func ParseClientCounterName(name string) (int64, error) {
	idStr := strings.TrimPrefix(name, clientCounterPrefix)
	return strconv.ParseInt(idStr, 16, 64)
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
