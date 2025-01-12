package alert

import (
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
)

const maximumNotificationThreshold int = 5
const maximumNotificationThresholdWindow time.Duration = 30 * time.Minute

type metricFilter struct {
	metricId          string
	lastNotifications utils.CircularBuffer[time.Time]
	spamCount         int
}

type alertFilters struct {
	filters map[string]*metricFilter
	input   <-chan metricIdWithMsg
	output  chan<- notifier.Message
}

type metricIdWithMsg struct {
	metricId string
	message  notifier.Message
}

func (mf *metricFilter) forwardToOutputIfAllowed(metricId string, msg notifier.Message, output chan<- notifier.Message) {
	now := time.Now()
	mf.lastNotifications.Push(now)
	oldestEntry := mf.lastNotifications.Front()

	if now.Sub(oldestEntry) <= maximumNotificationThresholdWindow && mf.lastNotifications.Full() {
		if mf.spamCount > 0 {
			logging.Debug("Metric %v: still spamming", mf.metricId)
		} else {
			output <- notifier.MakeMessage(notifier.Failure, "Notification spam detected (metricId: %v), skipping notifications", metricId)
			logging.Warning("Metric %v: spam detected", mf.metricId)
		}
		mf.spamCount++
	} else if mf.spamCount > 0 {
		logging.Info("Metric %v: end of spam (%v message lost)", mf.metricId, mf.spamCount)
		output <- notifier.MakeMessage(notifier.Recovery, "Notification spam has ended (metricId: %v, lost: %v)", metricId, mf.spamCount)
		mf.spamCount = 0
	}

	if mf.spamCount == 0 {
		logging.Debug("Forwarding : %v", msg)
		output <- msg
	} else {
		logging.Debug("Filtering : %v", msg)
	}
}

func (af *alertFilters) forwardToOutputIfAllowed(metricId string, msg notifier.Message) {
	if af.filters[metricId] == nil {
		af.filters[metricId] = &metricFilter{
			lastNotifications: utils.MakeCircularBuffer[time.Time](maximumNotificationThreshold),
		}
	}
	af.filters[metricId].forwardToOutputIfAllowed(metricId, msg, af.output)
}

func MakeAndRunAlertFilters(input <-chan metricIdWithMsg, output chan<- notifier.Message) {
	filtering := &alertFilters{
		filters: make(map[string]*metricFilter),
		input:   input,
		output:  output,
	}
	for inputMsg := range filtering.input {
		filtering.forwardToOutputIfAllowed(inputMsg.metricId, inputMsg.message)
	}
}
