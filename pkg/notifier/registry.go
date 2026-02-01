package notifier

import (
	"errors"
	"fmt"
)

var ErrNotifierNotFound = errors.New("notifier not found")

type Factory func(cfg Config) (Notifier, error)

var registry = make(map[string]Factory)

func RegisterNotifier(name string, factory Factory) {
	registry[name] = factory
}

func GetNotifier(name string) (Factory, error) {
	if factory, ok := registry[name]; ok {
		return factory, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrNotifierNotFound, name)
}
