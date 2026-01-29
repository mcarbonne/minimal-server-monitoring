package provider

import (
	"context"
	"fmt"
)

type Factory func(ctx context.Context, cfg Config) (Provider, error)

var registry = make(map[string]Factory)

func RegisterProvider(name string, factory Factory) {
	registry[name] = factory
}

func GetProvider(name string) (Factory, error) {
	if factory, ok := registry[name]; ok {
		return factory, nil
	}
	return nil, fmt.Errorf("provider '%s' not found", name)
}
