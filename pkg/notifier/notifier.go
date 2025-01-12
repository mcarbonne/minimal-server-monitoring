package notifier

import (
	"context"
	"fmt"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
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

func LoadAndRunNotifiers(ctx context.Context, machineName string, notifierCfgList map[string]Config, messageChan <-chan Message) {
	notifierList := make(map[string]Notifier, len(notifierCfgList))
	var err error
	for notifierName, notifierCfg := range notifierCfgList {
		if !utils.IsNameValid(notifierName) {
			logging.Fatal("Unable to setup notifier: forbidden characters in name '%v'", notifierName)
		}
		notifierList[notifierName], err = LoadNotifierFromConfig(notifierCfg)
		if err != nil {
			logging.Fatal("Unable to load notifier: %v", err)
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
