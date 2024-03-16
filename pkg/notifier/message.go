package notifier

import (
	"fmt"
	"strings"
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
	return Message{
		Type:    type_,
		Title:   strings.ToLower(fmt.Sprintf("%v", type_)),
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

	first := true
	for type_, cnt := range msgTypeMap {
		if first {
			title += fmt.Sprintf("%v: %v", type_, cnt)
		} else {
			title += fmt.Sprintf(", %v: %v", type_, cnt)
		}
		first = false
	}
	title = strings.ToLower(title)

	return Message{
		Type:    Aggregate,
		Title:   title,
		Message: description,
	}
}
