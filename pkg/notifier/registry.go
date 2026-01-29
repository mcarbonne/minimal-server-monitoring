package notifier

import (
	"fmt"
)

type Factory func(cfg Config) (Notifier, error)

var registry = make(map[string]Factory)

func RegisterNotifier(name string, factory Factory) {
	registry[name] = factory
}

func GetNotifier(name string) (Factory, error) {
	if factory, ok := registry[name]; ok {
		return factory, nil
	}
	return nil, fmt.Errorf("notifier '%s' not found", name)
}
