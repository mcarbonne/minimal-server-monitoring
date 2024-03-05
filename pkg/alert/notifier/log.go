package notifier

import "github.com/mcarbonne/minimal-server-monitoring/pkg/logging"

type Log struct {
}

func NewLog() Notifier {
	return &Log{}
}

func (l *Log) Send(message Message) error {
	logging.Info("New message:\nTitle : %v\nMessage: %v", message.Title, message.Message)
	return nil
}
