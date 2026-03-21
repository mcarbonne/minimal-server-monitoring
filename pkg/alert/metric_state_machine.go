package alert

import (
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper/customtypes"
)

type MetricStateMachine struct {
	healthyThreshold   uint // how many consecutive pass tests to mark metric as healthy (min 1)
	unhealthyThreshold uint // how many consecutive failed tests to mark metric as unhealthy (min 1)

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
		healthyThreshold:     max(1, healthyThreshold),
		unhealthyThreshold:   max(1, unhealthyThreshold),
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

func (msm *MetricStateMachine) shouldRemind(now time.Time) bool {

	if msm.reminderCounter < msm.failureReminderCount {
		if now.Sub(msm.lastFailureMessage) >= msm.failureReminder {
			return true
		}
	} else {
		nextReminder := nextDailyTime(msm.lastFailureMessage, msm.dailyReminder)
		if now.After(nextReminder) || now.Equal(nextReminder) {
			return true
		}
	}
	return false
}

func (msm *MetricStateMachine) Update(metricState provider.MetricState, now time.Time) *notifier.Message {

	isHealthyUpdate := metricState.Status == provider.Healthy || metricState.Status == provider.Removed

	if msm.isHealthy != isHealthyUpdate {
		msm.oppositeInARow++
	} else {
		msm.oppositeInARow = 0
	}

	// Special case: if service is removed, we want to clear the alert immediately
	if !msm.isHealthy && metricState.Status == provider.Removed {
		return msm.transitionToHealthy(metricState.Name, metricState.Description, "removed")
	}

	if msm.isHealthy {
		if msm.oppositeInARow >= msm.unhealthyThreshold {
			return msm.transitionToUnhealthy(metricState.Name, metricState.Description, now)
		}
	} else {
		if msm.oppositeInARow >= msm.healthyThreshold {
			return msm.transitionToHealthy(metricState.Name, metricState.Description, "recovered")
		} else if msm.shouldRemind(now) {
			msm.lastFailureMessage = now
			msm.reminderCounter++
			return makeMessage(notifier.Failure, "failed (reminder)", metricState.Name, metricState.Description)
		}

	}
	return nil
}

func (msm *MetricStateMachine) transitionToUnhealthy(name, description string, now time.Time) *notifier.Message {
	msm.isHealthy = false
	msm.oppositeInARow = 0
	msm.lastFailureMessage = now
	msm.reminderCounter = 0
	return makeMessage(notifier.Failure, "failed", name, description)
}

func (msm *MetricStateMachine) transitionToHealthy(name, description string, reason string) *notifier.Message {
	msm.isHealthy = true
	msm.oppositeInARow = 0
	msm.reminderCounter = 0
	msm.lastFailureMessage = time.Time{}
	return makeMessage(notifier.Recovery, reason, name, description)
}
