package notifier

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

type ConsoleNotifier struct {
}

func NewConsoleNotifier() Notifier {
	return &ConsoleNotifier{}
}

func (l *ConsoleNotifier) Send(message Message) error {
	logging.Info("New message:\nTitle : %v\nMessage: %v", message.Title, message.Message)
	return nil
}
