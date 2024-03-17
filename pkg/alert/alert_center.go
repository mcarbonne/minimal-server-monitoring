package alert

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping/provider"
)

func AlertCenter(alertCfg Config, scrapResultChan <-chan provider.ScrapeResult, notifyChan chan<- notifier.Message) {

	rawMessages := make(chan metricIdWithMsg)
	filteredMessages := make(chan notifier.Message)
	metricStateMachines := map[string]*MetricStateMachine{}

	//Step 1: convert scrape result to messages
	go func(outputChan chan<- metricIdWithMsg) {
		for scrapeResult := range scrapResultChan {
			for metricId, metricState := range scrapeResult.StateMap {
				if metricStateMachines[metricId] == nil {
					metricStateMachines[metricId] = MakeMetricStateMachine(metricId, alertCfg.HealthyThreshold, alertCfg.UnhealthyThreshold)
				}
				optMessage := metricStateMachines[metricId].Update(metricState)

				if optMessage != nil {
					outputChan <- metricIdWithMsg{
						metricId: metricId,
						message:  *optMessage}
				}
			}

			for _, msg := range scrapeResult.MessageList {
				outputChan <- metricIdWithMsg{
					metricId: msg.MetricID,
					message:  notifier.MakeMessage(notifier.Notification, msg.Description),
				}
			}
		}
	}(rawMessages)

	//Step 2: filter messages
	go func() {
		MakeAndRunAlertFilters(rawMessages, filteredMessages)
	}()

	//Step 3: group messages
	go func() {
		MakeAndRunAlertGrouping(filteredMessages, notifyChan)
	}()
}
