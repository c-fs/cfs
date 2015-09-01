package enforce

import (
	"sync"
	"time"

	"github.com/c-fs/cfs/stats"
	"github.com/codahale/metrics"
)

const flushIntervalSecond = 10

var (
	mu sync.Mutex
	// the number of client operations in the last checkpoint
	ops    = make(map[int64]int)
	quotas = make(map[int64]int)
)

func init() {
	updateOps()
	go func() {
		interval := time.Duration(flushIntervalSecond) * time.Second
		for _ = range time.Tick(interval) {
			updateOps()
		}
	}()
}

func updateOps() {
	mu.Lock()
	defer mu.Unlock()
	counters, _ := metrics.Snapshot()
	for k, v := range counters {
		id, err := stats.ParseClientCounterName(k)
		if err != nil {
			continue
		}
		ops[id] = int(v)
	}
}

func HasQuota(clientID int64) bool {
	if _, ok := quotas[clientID]; !ok {
		return true
	}

	counters, _ := metrics.Snapshot()
	nops := counters[stats.ClientCounterName(clientID)]

	mu.Lock()
	defer mu.Unlock()
	return int(nops)-ops[clientID] <= quotas[clientID]*flushIntervalSecond
}

func SetQuota(clientID int64, quota int) {
	mu.Lock()
	defer mu.Unlock()
	quotas[clientID] = quota
}
