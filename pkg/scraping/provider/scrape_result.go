package provider

import (
	"fmt"
)

type MetricState struct {
	MetricID    string
	IsHealthy   bool
	Description string
}

type MetricMessage struct {
	MetricID    string
	Description string
}

type ScrapeResultWrapper struct {
	prefix     string
	resultChan chan<- any
}

func MakeScrapeResultWrapper(prefix string, resultChan chan<- any) ScrapeResultWrapper {
	return ScrapeResultWrapper{
		prefix:     prefix,
		resultChan: resultChan,
	}
}

func (wrapper *ScrapeResultWrapper) getFullID(metricId string) string {
	return wrapper.prefix + "_" + metricId
}

func (wrapper *ScrapeResultWrapper) PushState(metricId string, isHealthy bool, description string, args ...any) {
	wrapper.resultChan <- MetricState{
		MetricID:    wrapper.getFullID(metricId),
		IsHealthy:   isHealthy,
		Description: fmt.Sprintf(description, args...),
	}
}

func (wrapper *ScrapeResultWrapper) PushFailure(metricId string, description string, args ...any) {
	wrapper.PushState(metricId, false, description, args...)
}

func (wrapper *ScrapeResultWrapper) PushOK(metricId string) {
	wrapper.PushState(metricId, true, "")
}

func (wrapper *ScrapeResultWrapper) PushMessage(metricId, description string, args ...any) {
	wrapper.resultChan <- MetricMessage{
		MetricID:    wrapper.getFullID(metricId),
		Description: fmt.Sprintf(description, args...),
	}
}
