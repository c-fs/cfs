package enforce

import (
	"testing"

	"github.com/c-fs/cfs/stats"
	"github.com/codahale/metrics"
)

func TestEnforceQuota(t *testing.T) {
	defer metrics.Reset()
	client := int64(0x1234)
	quota := 1
	SetQuota(client, quota)

	for i := 0; i < 5; i++ {
		updateOps()
		for k := 0; k < quota*flushIntervalSecond; k++ {
			metrics.Counter(stats.ClientCounterName(client)).Add()
			if !HasQuota(client) {
				t.Errorf("#%d.%d: unexpectedly out of quota", i, k)
			}
		}

		metrics.Counter(stats.ClientCounterName(client)).Add()
		if HasQuota(client) {
			t.Errorf("#%d: unexpectedly have quota", i)
		}
	}
}
