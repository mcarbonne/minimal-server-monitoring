package alert

import (
	"context"
	"fmt"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping/provider"
)

func AlertCenter(ctx context.Context, alertCfg Config, scrapResultChan <-chan any, notifyChan chan<- notifier.Message) {

	rawNotifications := make(chan metricIdWithMsg)
	filteredNotifications := make(chan notifier.Message)
	metricStateMachines := map[string]*MetricStateMachine{}

	//Step 1: convert scrape result to messages
	go func(outputChan chan<- metricIdWithMsg) {
		for scrapeResult := range scrapResultChan {
			switch element := scrapeResult.(type) {
			case provider.MetricMessage:
				outputChan <- metricIdWithMsg{
					metricId: element.MetricID,
					message: notifier.MakeMessage(notifier.Notification,
						fmt.Sprintf("%v: %v", element.Name, element.Description)),
				}
			case provider.MetricState:
				if metricStateMachines[element.MetricID] == nil {
					metricStateMachines[element.MetricID] = MakeMetricStateMachine(alertCfg.HealthyThreshold, alertCfg.UnhealthyThreshold)
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
	}(rawNotifications)

	//Step 2: filter messages
	go func() {
		MakeAndRunAlertFilters(rawNotifications, filteredNotifications)
	}()

	//Step 3: group messages
	go func() {
		MakeAndRunAlertGrouping(filteredNotifications, notifyChan)
	}()
}
