package notifier

import (
	"fmt"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

type Notifier interface {
	Send(message Message) error
}

func LoadNotifierFromConfig(cfg Config) (Notifier, error) {
	switch cfg.Type {
	case "shoutrrr":
		return NewShoutrrr(cfg.Params)
	case "console":
		return NewConsoleNotifier(), nil
	default:
		return nil, fmt.Errorf("illegal notifier type: %v", cfg.Type)
	}
}

func LoadAndRunNotifiers(machineName string, notifierCfgList []Config, messageChan <-chan Message) {
	notifierList := make([]Notifier, len(notifierCfgList))
	var err error
	for i, notifierCfg := range notifierCfgList {
		notifierList[i], err = LoadNotifierFromConfig(notifierCfg)
		if err != nil {
			logging.Fatal("Unable to load notifier: %v", err)
		}
	}
	for msg := range messageChan {
		msg.Title = machineName + " " + msg.Title
		for _, notifier := range notifierList {
			notifier.Send(msg)
		}
	}
}
