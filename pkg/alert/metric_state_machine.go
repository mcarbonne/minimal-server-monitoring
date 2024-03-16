package alert

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
)

type MetricStateMachine struct {
	metricId           string
	healthyThreshold   uint // how many consecutived pass tests to mark metric as healthy, 0 means metric is healthy on first fail
	unhealthyThreshold uint // how many consecutive failed tests to mark metric as unhealthy, 0 means metric is unhealthy on first fail

	isHealthy      bool
	oppositeInARow uint
}

func MakeMetricStateMachine(metricId string, healthyThreshold, unhealthyThreshold uint) *MetricStateMachine {
	return &MetricStateMachine{
		metricId:           metricId,
		healthyThreshold:   healthyThreshold,
		unhealthyThreshold: unhealthyThreshold,
		isHealthy:          true,
		oppositeInARow:     0,
	}
}

func (msm *MetricStateMachine) Update(metricState provider.MetricState) *notifier.Message {

	var msg *notifier.Message

	if msm.isHealthy != metricState.IsHealthy {
		msm.oppositeInARow++
	} else {
		msm.oppositeInARow = 0
	}

	if msm.isHealthy {
		if msm.oppositeInARow >= msm.unhealthyThreshold {
			msm.isHealthy = false
			msg = utils.Ptr(notifier.MakeMessage(notifier.Failure, "metric %v failed: %v", msm.metricId, metricState.Description))
		}
	} else {
		if msm.oppositeInARow >= msm.healthyThreshold {
			msm.isHealthy = true
			msg = utils.Ptr(notifier.MakeMessage(notifier.OK, "metric %v is OK", msm.metricId))
		}

	}
	return msg
}
