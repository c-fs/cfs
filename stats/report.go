package stats

import (
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/codahale/metrics"
	influx "github.com/influxdb/influxdb/client"
	"github.com/qiniu/log"
)

type Sinker interface {
	Sink(counters map[string]uint64, dur time.Duration)
}

// If sinker is nil, it will use default sinker.
// Default sinker just print counters by log.
func Report(sinker Sinker, interval time.Duration) {
	once.Do(func() { go report(sinker, interval) })
}

//-----------------------------------------------------------------------------

var once sync.Once

func report(sinker Sinker, interval time.Duration) {

	if sinker == nil {
		sinker = defaultSinker
	}

	var (
		lastTime        = time.Now()
		lastSnapshot, _ = metrics.Snapshot()
	)

	for _ = range time.Tick(interval) {
		var (
			now           = time.Now()
			counters, _   = metrics.Snapshot()
			deltaCounters = make(map[string]uint64)
			dur           = now.Sub(lastTime)
		)
		for m, n := range counters {
			o, _ := lastSnapshot[m]
			if delta := n - o; delta != 0 {
				deltaCounters[m] = delta
			}
		}
		lastTime, lastSnapshot = now, counters
		sinker.Sink(deltaCounters, dur)
	}
}

//-----------------------------------------------------------------------------

var defaultSinker = logSinker{log.Std}

type logSinker struct {
	*log.Logger
}

func (l logSinker) Sink(counters map[string]uint64, dur time.Duration) {

	for m, n := range counters {
		l.Infof("stats: %s %.1f ops", m, float64(n)/dur.Seconds())
	}
}

//-----------------------------------------------------------------------------

type InfluxConfig struct {
	Address  string // host:port
	Username string
	Password string
	Database string
}

func NewInfluxSinker(conf InfluxConfig) (Sinker, error) {

	u, err := url.Parse("http://" + conf.Address)
	if err != nil {
		return nil, err
	}
	cli, err := influx.NewClient(influx.Config{
		URL:      *u,
		Username: conf.Username,
		Password: conf.Password,
	})

	if err != nil {
		return nil, err
	}

	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &influxSinker{
		influx:   cli,
		hostName: host,
		database: conf.Database,
	}, nil
}

type influxSinker struct {
	influx   *influx.Client
	database string
	hostName string
}

func (s *influxSinker) Sink(counters map[string]uint64, dur time.Duration) {

	var (
		points = make([]influx.Point, 0, len(counters))
		now    = time.Now()
	)

	for m, n := range counters {
		var (
			measurement string
			tags        = map[string]string{"host": s.hostName}
		)
		if strings.HasPrefix(m, clientCounterPrefix) {
			measurement = "client_ops"
			tags["id"] = strings.TrimPrefix(m, clientCounterPrefix)
		} else {
			measurement = "server_ops"
			fields := strings.SplitN(m, "_", 2)
			if len(fields) != 2 {
				log.Errorf("influx sinker found unepxected metric: %s", m)
				continue
			}
			tags["disk"] = fields[0]
			tags["op"] = fields[1]
		}
		p := influx.Point{
			Measurement: measurement,
			Time:        now,
			Tags:        tags,
			Fields:      map[string]interface{}{"value": n},
		}
		points = append(points, p)
	}

	if _, err := s.influx.Write(influx.BatchPoints{
		Database: s.database,
		Points:   points,
	}); err != nil {
		log.Error("influx write failed -", err)
	}
}
