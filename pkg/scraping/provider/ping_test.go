package provider

import (
	"context"
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"gotest.tools/v3/assert"
)

type mockPinger struct {
	PingFunc func(target string) bool
}

func (m *mockPinger) Ping(target string) bool {
	if m.PingFunc != nil {
		return m.PingFunc(target)
	}
	return false
}

func TestPing(t *testing.T) {
	mock := &mockPinger{}
	provider := &ProviderPing{
		pinger:     mock,
		Targets:    []string{"1.1.1.1", "bad.host"},
		RetryCount: 3,
	}

	resultChan := make(chan any, 10)
	wrapper := MakeScrapeResultWrapper("ping", resultChan)

	// Scenario: 1.1.1.1 OK, bad.host FAIL
	callCount := make(map[string]int)
	mock.PingFunc = func(target string) bool {
		callCount[target]++
		return target == "1.1.1.1"
	}

	taskList := provider.GetUpdateTaskList(context.Background(), &wrapper, storage.NewMemoryStorage())

	// There are 2 tasks (one per target)
	assert.Equal(t, 2, len(taskList))

	// Execute tasks
	for _, task := range taskList {
		task()
	}

	// Verify 1.1.1.1
	metricA := waitForMetricState(t, resultChan, "ping_ping_1.1.1.1")
	assert.Equal(t, true, metricA.IsHealthy)

	// Verify bad.host
	metricB := waitForMetricState(t, resultChan, "ping_ping_bad.host")
	assert.Equal(t, false, metricB.IsHealthy)
	assert.Equal(t, "unreachable", metricB.Description)

	// Verify retry logic was used
	// 1.1.1.1 -> 1 call (success immediately)
	// bad.host -> 3 calls (fail 3 times)
	assert.Equal(t, 1, callCount["1.1.1.1"])
	assert.Equal(t, 3, callCount["bad.host"])
}
