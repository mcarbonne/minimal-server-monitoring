package stats

import (
	"fmt"
	"time"
)

type WindowCollector[T any] struct {
	data           []TimestampedData[T]
	windowDuration time.Duration
}

func MakeWindowCollector[T any](windowDuration time.Duration) WindowCollector[T] {
	return WindowCollector[T]{
		data:           make([]TimestampedData[T], 0),
		windowDuration: windowDuration,
	}
}

func (collector *WindowCollector[T]) String() string {
	return fmt.Sprintf("%v", collector.data)
}

func (collector *WindowCollector[T]) AddNew(data T) {
	now := time.Now()
	collector.removeOldEntries(now)
	collector.data = append(collector.data, TimestampedData[T]{now, data})
}

func (collector *WindowCollector[T]) First() TimestampedData[T] {
	return collector.data[0]
}

func (collector *WindowCollector[T]) Last() TimestampedData[T] {
	return collector.data[len(collector.data)-1]
}

func (collector *WindowCollector[T]) Count() int {
	return len(collector.data)
}

func (collector *WindowCollector[T]) removeOldEntries(now time.Time) {
	var i int
	for i = range collector.data {
		if now.Sub(collector.data[i].Timestamp) < collector.windowDuration {
			break
		}
	}
	collector.data = collector.data[i:]
}
