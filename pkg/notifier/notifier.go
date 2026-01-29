package notifier

import (
	"context"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
)

type Notifier interface {
	Send(message Message) error
}

func LoadNotifierFromConfig(cfg Config) (Notifier, error) {
	factory, err := GetNotifier(cfg.Type)
	if err != nil {
		return nil, err
	}
	return factory(cfg)
}

func LoadAndRunNotifiers(ctx context.Context, machineName string, notifierCfgList map[string]Config, messageChan <-chan Message) {
	notifierList := make(map[string]Notifier, len(notifierCfgList))
	var err error
	for notifierName, notifierCfg := range notifierCfgList {
		if !utils.IsNameValid(notifierName) {
			logging.Error("Unable to setup notifier: forbidden characters in name '%v'", notifierName)
			continue
		}
		notifierList[notifierName], err = LoadNotifierFromConfig(notifierCfg)
		if err != nil {
			logging.Error("Unable to load notifier: %v", err)
			continue
		}
	}

mainloop:
	for {
		select {
		case <-ctx.Done():
			logging.Info("Exiting notifier")
			break mainloop
		case msg := <-messageChan:
			msg.Title = machineName + " " + msg.Title
			for _, notifier := range notifierList {
				err = notifier.Send(msg)
				if err != nil {
					logging.Error("Failed to notify: %v", err)
				}
			}
		}
	}
}
