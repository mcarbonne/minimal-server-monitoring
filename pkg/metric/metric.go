package metric

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

type Metric struct {
	name     string
	provider Provider
	policy   Policy
	state    State
}

func MakeMetric(name string, provider Provider, policy Policy) *Metric {
	return &Metric{name: name,
		provider: provider,
		policy:   policy,
	}
}

func (m *Metric) Update(storage storage.Storager, eventChannel chan<- Event) {
	resultList := m.provider.Update(storage)
	newState := StateHealthy
	for _, v := range resultList.list {
		if v.EventType == TypeFailure {
			newState = StateUnhealthy
		}
		eventChannel <- MakeEvent(m.name+"_"+v.Subtopic, v.EventType, v.Description)
	}

	if newState == StateHealthy && m.state == StateUnhealthy {
		eventChannel <- MakeEvent(m.name+"_health", TypeInfo, "Metric %v is healthy again", m.name)
	}
	m.state = newState
}
