package stats_test

import (
	"testing"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/stats"
	"gotest.tools/v3/assert"
)

func TestWindowCollector(t *testing.T) {
	collector := stats.MakeWindowCollector[int](1 * time.Second)
	for i := range 5 {
		collector.AddNew(i)
		time.Sleep(400 * time.Millisecond)
	}
	assert.Equal(t, collector.First().Data, 2)
	assert.Equal(t, collector.Last().Data, 4)
}
