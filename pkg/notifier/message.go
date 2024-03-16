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
	OK
	Aggregate
)

type Message struct {
	Type    MessageType
	Title   string
	Message string
}

func MakeMessage(type_ MessageType, descriptionFormat string, args ...any) Message {
	hostname, err := os.Hostname()
	if err != nil {
		logging.Warning("Unable to get hostname")
	}

	return Message{
		Type:    type_,
		Title:   fmt.Sprintf("%v %v", hostname, type_),
		Message: fmt.Sprintf(descriptionFormat, args...),
	}
}

func MakeAggregatedMessage(msgList []Message) Message {
	msgTypeMap := make(map[MessageType]int)
	var title, description string
	for _, msg := range msgList {
		msgTypeMap[msg.Type]++
		description += " - " + msg.Message + "\n"
	}

	hostname, err := os.Hostname()
	if err != nil {
		logging.Warning("Unable to get hostname")
	} else {
		title = hostname
	}

	for type_, cnt := range msgTypeMap {
		title += fmt.Sprintf(" %v: %v", type_, cnt)
	}

	return Message{
		Type:    Aggregate,
		Title:   title,
		Message: description,
	}
}
