package provider

import (
	"context"
	"fmt"
)

type Config struct {
	Type           string         `json:"type"`
	ScrapeInterval uint           `json:"scrape_interval" default:"120"` // scrape interval in seconds
	Params         map[string]any `json:"params" default:"{}"`           // extra parameters
}

func LoadProviderFromConfig(ctx context.Context, cfg Config) (Provider, error) {
	switch cfg.Type {
	case "container":
		return NewProviderContainer()
	case "ping":
		return NewProviderPing(cfg.Params)
	case "filesystemusage":
		return NewProviderFileSystemUsage(cfg.Params)
	default:
		return nil, fmt.Errorf("illegal provider type: %v", cfg.Type)
	}
}
