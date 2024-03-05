package metric

import "fmt"

type ProviderResult struct {
	EventType   EventType
	Subtopic    string
	Description string
}

type ProviderResultList struct {
	list []ProviderResult
}

func MakeProviderResultList() ProviderResultList {
	return ProviderResultList{
		list: []ProviderResult{},
	}
}

func (list *ProviderResultList) Push(eventType EventType, subtopic, description string, args ...any) {
	list.list = append(list.list, ProviderResult{
		EventType:   eventType,
		Subtopic:    subtopic,
		Description: fmt.Sprintf(description, args...),
	})
}
