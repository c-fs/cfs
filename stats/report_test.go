package stats

import (
	"sort"
	"testing"
	"time"
)

func TestReport(t *testing.T) {
	ch := make(chan counter)
	interval := 10 * time.Millisecond

	Report(csinker(ch), interval)
	time.Sleep(interval)

	Counter("cfs0", "read").Client(0x1234).Add()
	time.Sleep(2 * interval)
	(<-ch).Assert(t, "cfs0_read", 1)
	(<-ch).Assert(t, "client_1234", 1)

	Counter("cfs0", "write").Client(0x1234).Add()
	Counter("cfs1", "read").Client(0x1234).Add()
	Counter("cfs1", "read").Client(0x1234).Add()
	time.Sleep(2 * interval)
	(<-ch).Assert(t, "cfs0_write", 1)
	(<-ch).Assert(t, "cfs1_read", 2)
	(<-ch).Assert(t, "client_1234", 3)
}

type counter struct {
	Name string
	Val  uint64
}

func (c counter) Assert(t *testing.T, name string, val uint64) {

	if c.Name != name || c.Val != val {
		t.Fatalf("expect(%s %d), actual(%s %d)", name, val, c.Name, c.Val)
	}
}

type countersSort []counter

func (c countersSort) Len() int           { return len(c) }
func (c countersSort) Less(a, b int) bool { return c[a].Name < c[b].Name }
func (c countersSort) Swap(a, b int)      { c[a], c[b] = c[b], c[a] }

type csinker chan counter

func (m csinker) Sink(counters map[string]uint64, dur time.Duration) {

	cs := make([]counter, 0, len(counters))
	for c, n := range counters {
		cs = append(cs, counter{c, n})
	}

	sort.Sort(countersSort(cs))

	for _, c := range cs {
		m <- c
	}
}
