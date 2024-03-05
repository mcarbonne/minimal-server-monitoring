package metric

import (
	"fmt"
	"time"
)

//go:generate stringer -type EventType -linecomment
type EventType uint

const (
	TypeUndefined EventType = iota // Undefined
	TypeInfo                       // Info
	TypeFailure                    // Failure
)

type Event struct {
	Timestamp   time.Time
	Topic       string
	EventType   EventType
	Description string
}

func MakeEvent(topic string, eventType EventType, descriptionFormat string, args ...any) Event {
	return Event{
		Timestamp:   time.Now().Round(0),
		Topic:       topic,
		EventType:   eventType,
		Description: fmt.Sprintf(descriptionFormat, args...)}
}

func (evt *Event) String() string {
	return fmt.Sprintf("at: %v, topic: %v, type: %v, description: %v", evt.Timestamp, evt.Topic, evt.EventType, evt.Description)
}
