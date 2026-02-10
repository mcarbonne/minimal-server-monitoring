package provider

import (
	"context"
	"testing"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
)

func waitForMetricState(t *testing.T, ch chan any, metricID string) MetricState {
	return waitForType(t, ch, metricID, func(m MetricState) string { return m.MetricID })
}

func waitForMessage(t *testing.T, ch chan any, metricID string) MetricMessage {
	return waitForType(t, ch, metricID, func(m MetricMessage) string { return m.MetricID })
}

func waitForType[T any](t *testing.T, ch chan any, expectedID string, getID func(T) string) T {
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case msg := <-ch:
			if typedMsg, ok := msg.(T); ok {
				if getID(typedMsg) == expectedID {
					return typedMsg
				}
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for metric %s", expectedID)
		}
	}
}

func drainChannel(ch chan any) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func getAndExecuteTaskList(provider Provider, ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) {
	updateTaskList := provider.GetUpdateTaskList(ctx, resultWrapper, storage)

	for _, updateTask := range updateTaskList {
		updateTask()
	}
}
