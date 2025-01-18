package provider

import (
	"fmt"
)

type MetricState struct {
	MetricID    string
	Name        string
	IsHealthy   bool
	Description string
}

type MetricMessage struct {
	MetricID    string
	Name        string
	Description string
}

type ScrapeResultWrapper struct {
	prefix     string
	resultChan chan<- any
}

type MetricWrapper struct {
	metricID      string
	name          string
	resultWrapper *ScrapeResultWrapper
}

func MakeScrapeResultWrapper(prefix string, resultChan chan<- any) ScrapeResultWrapper {
	return ScrapeResultWrapper{
		prefix:     prefix,
		resultChan: resultChan,
	}
}

func (wrapper *ScrapeResultWrapper) Metric(metricId, name string) MetricWrapper {
	return MetricWrapper{
		metricID:      metricId,
		name:          name,
		resultWrapper: wrapper,
	}
}

func (wrapper *ScrapeResultWrapper) getFullID(metricId string) string {
	return wrapper.prefix + "_" + metricId
}

func (wrapper *ScrapeResultWrapper) pushState(metricId, name string, isHealthy bool, description string, args ...any) {
	wrapper.resultChan <- MetricState{
		MetricID:    wrapper.getFullID(metricId),
		Name:        name,
		IsHealthy:   isHealthy,
		Description: fmt.Sprintf(description, args...),
	}
}

func (wrapper *MetricWrapper) PushFailure(description string, args ...any) {
	wrapper.resultWrapper.pushState(wrapper.metricID, wrapper.name, false, description, args...)
}

func (wrapper *MetricWrapper) PushOK(description string, args ...any) {
	wrapper.resultWrapper.pushState(wrapper.metricID, wrapper.name, true, description, args...)
}

func (wrapper *MetricWrapper) PushMessage(description string, args ...any) {
	wrapper.resultWrapper.resultChan <- MetricMessage{
		MetricID:    wrapper.resultWrapper.getFullID(wrapper.metricID),
		Name:        wrapper.name,
		Description: fmt.Sprintf(description, args...),
	}
}
