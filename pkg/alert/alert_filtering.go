package alert

import (
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/alert/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

const maximumNotificationThreshold int = 5
const maximumNotificationThresholdWindow time.Duration = 30 * time.Minute
const alertFilteringTopic = "alert_filtering"

type topicFilter struct {
	currentIndex                    int
	lastNotificationsCircularBuffer [maximumNotificationThreshold]time.Time
	spamCount                       int
}

type alertFilters struct {
	filters map[string]*topicFilter
}

func MakeAlertFilters() *alertFilters {
	return &alertFilters{
		filters: make(map[string]*topicFilter),
	}
}

func (tf *topicFilter) sendIfAllowed(msg notifier.Message, notify func(notifier.Message)) {
	tf.lastNotificationsCircularBuffer[tf.currentIndex] = time.Now()
	tf.currentIndex = (tf.currentIndex + 1) % maximumNotificationThreshold
	oldestEntry := tf.lastNotificationsCircularBuffer[tf.currentIndex]

	if time.Since(oldestEntry) <= maximumNotificationThresholdWindow {
		if tf.spamCount > 0 {
			logging.Debug("Topic %v: still spamming", msg.Topic)
		} else {
			notify(notifier.MakeMessage(alertFilteringTopic, "spam detected", "Notification spam detected (topic: %v), skipping notifications", msg.Topic))
			logging.Warning("Topic %v: spam detected", msg.Topic)
		}
		tf.spamCount++
	} else if tf.spamCount > 0 {
		logging.Info("Topic %v: end of spam (%v message lost)", msg.Topic, tf.spamCount)
		notify(notifier.MakeMessage(alertFilteringTopic, "end of spam", "Notification spam has ended (topic: %v), %v message lost", msg.Topic, tf.spamCount))
		tf.spamCount = 0
	}

	if tf.spamCount == 0 {
		logging.Debug("Sending : %v", msg)
		notify(msg)
	} else {
		logging.Debug("Filtering : %v", msg)
	}
}

func (af *alertFilters) sendIfAllowed(msg notifier.Message, notify func(notifier.Message)) {
	if af.filters[msg.Topic] == nil {
		af.filters[msg.Topic] = &topicFilter{}
	}
	af.filters[msg.Topic].sendIfAllowed(msg, notify)
}
