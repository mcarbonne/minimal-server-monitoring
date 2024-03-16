package notifier

import (
	"fmt"
	"os"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

//go:generate stringer -type MessageType
type MessageType uint

const (
	Undefined MessageType = iota
	Notification
	Failure
	Recovery
)

type Message struct {
	MetricId string
	Title    string
	Message  string
}

func MakeMessage(type_ MessageType, metricId, descriptionFormat string, args ...any) Message {
	hostname, err := os.Hostname()
	if err != nil {
		logging.Warning("Unable to get hostname")
	}

	return Message{
		MetricId: metricId,
		Title:    fmt.Sprintf("%v %v", hostname, type_),
		Message:  fmt.Sprintf(descriptionFormat, args...),
	}
}
