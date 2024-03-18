package provider

import (
	"context"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

type Config struct {
	Type           string         `json:"type"`
	ScrapeInterval uint           `json:"scrape_interval" default:"120"` // scrape interval in seconds
	Params         map[string]any `json:"params" default:"{}"`           // extra parameters
}

func LoadProviderFromConfig(ctx context.Context, cfg Config) Provider {
	switch cfg.Type {
	case "container":
		return NewProviderContainer()
	case "ping":
		return NewProviderPing(cfg.Params)
	case "filesystemusage":
		return NewProviderFileSystemUsage(cfg.Params)
	default:
		logging.Fatal("Illegal provider type: %v", cfg.Type)
		return nil
	}
}
