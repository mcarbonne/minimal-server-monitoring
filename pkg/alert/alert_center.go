package alert

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/alert/notifier"
)

func sendToAll(notifierList []notifier.Notifier, msg notifier.Message) {
	for _, notifierInst := range notifierList {
		notifierInst.Send(msg)
	}
}

func AlertCenter(notifierCfgList []notifier.Config, msgChan <-chan notifier.Message) {
	notifiersList := make([]notifier.Notifier, len(notifierCfgList))
	for i, notifierCfg := range notifierCfgList {
		notifiersList[i] = notifier.LoadNotifierFromConfig(notifierCfg)
	}

	filtering := MakeAlertFilters()

	for msg := range msgChan {
		// filter to avoid notification spam
		filtering.sendIfAllowed(msg, func(msg notifier.Message) { sendToAll(notifiersList, msg) })
	}
}
