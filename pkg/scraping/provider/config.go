package provider

import (
	"context"
	"fmt"
	"time"
)

type Config struct {
	Type           string         `json:"type"`
	ScrapeInterval time.Duration  `json:"scrape_interval" default:"120s"` // scrape interval
	Params         map[string]any `json:"params" default:"{}"`            // extra parameters
}

func LoadProviderFromConfig(ctx context.Context, cfg Config) (Provider, error) {
	switch cfg.Type {
	case "container":
		return NewProviderContainer()
	case "ping":
		return NewProviderPing(cfg.Params)
	case "systemd":
		return NewProviderSystemd(ctx, cfg.Params)
	case "filesystemusage":
		return NewProviderFileSystemUsage(cfg.Params)
	default:
		return nil, fmt.Errorf("illegal provider type: %v", cfg.Type)
	}
}
