package notifier

import (
	"fmt"
	"os"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

type Message struct {
	Topic   string // internal topic, used for filtering
	Title   string
	Message string
}

func MakeMessage(topic, title, descriptionFormat string, args ...any) Message {
	hostname, err := os.Hostname()
	if err != nil {
		logging.Warning("Unable to get hostname")
	}

	return Message{
		Topic:   topic,
		Title:   fmt.Sprintf("%v %v", hostname, title),
		Message: fmt.Sprintf(descriptionFormat, args...),
	}
}
