package alert

import (
	"context"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping/provider"
)

func AlertCenter(ctx context.Context, alertCfg Config, scrapResultChan <-chan any, notifyChan chan<- notifier.Message) {

	rawMessages := make(chan metricIdWithMsg)
	filteredMessages := make(chan notifier.Message)
	metricStateMachines := map[string]*MetricStateMachine{}

	//Step 1: convert scrape result to messages
	go func(outputChan chan<- metricIdWithMsg) {
		for scrapeResult := range scrapResultChan {
			switch element := scrapeResult.(type) {
			case provider.MetricMessage:
				outputChan <- metricIdWithMsg{
					metricId: element.MetricID,
					message:  notifier.MakeMessage(notifier.Notification, element.Description),
				}
			case provider.MetricState:
				if metricStateMachines[element.MetricID] == nil {
					metricStateMachines[element.MetricID] = MakeMetricStateMachine(element.MetricID, alertCfg.HealthyThreshold, alertCfg.UnhealthyThreshold)
				}
				optMessage := metricStateMachines[element.MetricID].Update(element)

				if optMessage != nil {
					outputChan <- metricIdWithMsg{
						metricId: element.MetricID,
						message:  *optMessage}
				}
			default:
				logging.Warning("Unsupported element received: %v", element)
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
