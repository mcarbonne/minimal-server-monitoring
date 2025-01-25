package alert

import (
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
)

type MetricStateMachine struct {
	healthyThreshold   uint // how many consecutived pass tests to mark metric as healthy, 0 means metric is healthy on first fail
	unhealthyThreshold uint // how many consecutive failed tests to mark metric as unhealthy, 0 means metric is unhealthy on first fail

	isHealthy      bool
	oppositeInARow uint

	failureReminder    time.Duration
	lastFailureMessage time.Time
}

func MakeMetricStateMachine(healthyThreshold, unhealthyThreshold uint, failureReminderDelay time.Duration) *MetricStateMachine {
	return &MetricStateMachine{
		healthyThreshold:   healthyThreshold,
		unhealthyThreshold: unhealthyThreshold,
		isHealthy:          true,
		oppositeInARow:     0,
		failureReminder:    failureReminderDelay,
		lastFailureMessage: time.Time{},
	}
}

func makeMessage(msgType notifier.MessageType, what, name, description string) *notifier.Message {
	if description == "" {
		return utils.Ptr(notifier.MakeMessage(msgType, "%v %v", name, what))
	} else {
		return utils.Ptr(notifier.MakeMessage(msgType, "%v %v: %v", name, what, description))
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
			msm.lastFailureMessage = time.Now()
			msg = makeMessage(notifier.Failure, "failed", metricState.Name, metricState.Description)
		}
	} else {
		if msm.oppositeInARow >= msm.healthyThreshold {
			msm.isHealthy = true
			msg = makeMessage(notifier.Recovery, "recovered", metricState.Name, metricState.Description)
		} else if time.Now().Sub(msm.lastFailureMessage) >= msm.failureReminder {
			msm.lastFailureMessage = time.Now()
			msg = makeMessage(notifier.Failure, "failed (reminder)", metricState.Name, metricState.Description)
		}

	}
	return msg
}
