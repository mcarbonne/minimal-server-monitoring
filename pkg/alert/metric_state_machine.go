package alert

import (
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper/customtypes"
)

type MetricStateMachine struct {
	healthyThreshold   uint // how many consecutived pass tests to mark metric as healthy, 0 means metric is healthy on first fail
	unhealthyThreshold uint // how many consecutive failed tests to mark metric as unhealthy, 0 means metric is unhealthy on first fail

	isHealthy      bool
	oppositeInARow uint

	failureReminder      time.Duration
	failureReminderCount uint
	dailyReminder        customtypes.TimeOfDay

	lastFailureMessage time.Time
	reminderCounter    uint
}

func MakeMetricStateMachine(healthyThreshold, unhealthyThreshold uint, failureReminderDelay time.Duration, failureReminderCount uint, dailyReminder customtypes.TimeOfDay) *MetricStateMachine {
	return &MetricStateMachine{
		healthyThreshold:     healthyThreshold,
		unhealthyThreshold:   unhealthyThreshold,
		isHealthy:            true,
		oppositeInARow:       0,
		failureReminder:      failureReminderDelay,
		failureReminderCount: failureReminderCount,
		dailyReminder:        dailyReminder,
		lastFailureMessage:   time.Time{},
		reminderCounter:      0,
	}
}

func makeMessage(msgType notifier.MessageType, what, name, description string) *notifier.Message {
	if description == "" {
		return utils.Ptr(notifier.MakeMessage(msgType, "%v %v", name, what))
	} else {
		return utils.Ptr(notifier.MakeMessage(msgType, "%v %v: %v", name, what, description))
	}
}

func nextDailyTime(after time.Time, target customtypes.TimeOfDay) time.Time {
	candidate := time.Date(after.Year(), after.Month(), after.Day(), target.Hour, target.Minute, 0, 0, after.Location())
	if candidate.After(after) {
		return candidate
	}
	return candidate.Add(24 * time.Hour)
}

func (msm *MetricStateMachine) Update(metricState provider.MetricState, now time.Time) *notifier.Message {

	var msg *notifier.Message

	if msm.isHealthy != metricState.IsHealthy {
		msm.oppositeInARow++
	} else {
		msm.oppositeInARow = 0
	}

	if msm.isHealthy {
		if msm.oppositeInARow >= msm.unhealthyThreshold {
			msm.isHealthy = false
			msm.lastFailureMessage = now
			msm.reminderCounter = 0
			msg = makeMessage(notifier.Failure, "failed", metricState.Name, metricState.Description)
		}
	} else {
		if msm.oppositeInARow >= msm.healthyThreshold {
			msm.isHealthy = true
			msg = makeMessage(notifier.Recovery, "recovered", metricState.Name, metricState.Description)
		} else {
			shouldRemind := false
			if msm.reminderCounter < msm.failureReminderCount {
				if now.Sub(msm.lastFailureMessage) >= msm.failureReminder {
					shouldRemind = true
				}
			} else {
				nextReminder := nextDailyTime(msm.lastFailureMessage, msm.dailyReminder)
				if now.After(nextReminder) || now.Equal(nextReminder) {
					shouldRemind = true
				}
			}

			if shouldRemind {
				msm.lastFailureMessage = now
				msm.reminderCounter++
				msg = makeMessage(notifier.Failure, "failed (reminder)", metricState.Name, metricState.Description)
			}
		}

	}
	return msg
}
