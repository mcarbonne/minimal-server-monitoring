package provider

import (
	"fmt"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

type MetricState struct {
	IsHealthy   bool
	Description string
}

type MetricMessage struct {
	MetricID    string
	Description string
}

type ScrapeResult struct {
	prefix         string
	MetricStateMap map[string]MetricState // only one state per metric allowed
	MessageList    []MetricMessage        // list of messages, multiple messages per metric allowed
}

func MakeScrapeResult(prefix string) ScrapeResult {
	return ScrapeResult{
		prefix:         prefix,
		MetricStateMap: map[string]MetricState{},
		MessageList:    []MetricMessage{},
	}
}

func (list *ScrapeResult) getFullID(metricId string) string {
	return list.prefix + "_" + metricId
}

func (list *ScrapeResult) PushState(metricId string, isHealthy bool, description string, args ...any) {
	fullID := list.getFullID(metricId)
	_, exists := list.MetricStateMap[fullID]
	if exists {
		logging.Fatal("%v already exists", metricId)
	} else {
		list.MetricStateMap[fullID] = MetricState{
			IsHealthy:   isHealthy,
			Description: fmt.Sprintf(description, args...),
		}
	}
}

func (list *ScrapeResult) PushFailure(metricId string, description string, args ...any) {
	list.PushState(metricId, false, description, args...)
}

func (list *ScrapeResult) PushOK(metricId string) {
	list.PushState(metricId, true, "")
}

func (list *ScrapeResult) PushMessage(metricId, description string, args ...any) {
	list.MessageList = append(list.MessageList, MetricMessage{
		MetricID:    list.getFullID(metricId),
		Description: fmt.Sprintf(description, args...),
	})
}
