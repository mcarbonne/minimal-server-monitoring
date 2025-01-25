package alert

import (
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/notifier"
)

const maxAggregatedMessages int = 10

func MakeAndRunAlertGrouping(cfg GroupingConfig, input <-chan notifier.Message, output chan<- notifier.Message) {
	currentGroup := make([]notifier.Message, 0, maxAggregatedMessages)
	nextDeadline := time.Now().Add(cfg.Window)

	sendAndFlush := func() {
		logging.Info("Sending a grouped message of size %v", len(currentGroup))
		output <- notifier.MakeAggregatedMessage(currentGroup)
		currentGroup = currentGroup[:0]
	}

	for {
		select {
		case msg := <-input:
			if len(currentGroup) == 0 {
				nextDeadline = time.Now().Add(cfg.Window)
			}
			currentGroup = append(currentGroup, msg)
		case <-time.After(time.Until(nextDeadline)):
			if len(currentGroup) > 0 {
				sendAndFlush()
			} else {
				nextDeadline = time.Now().Add(cfg.Window)
			}
		}

		if len(currentGroup) >= maxAggregatedMessages {
			sendAndFlush()
		}
	}
}
