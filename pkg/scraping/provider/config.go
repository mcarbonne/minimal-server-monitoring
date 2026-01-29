package provider

import (
	"context"
	"time"
)

type Config struct {
	Type           string         `json:"type"`
	ScrapeInterval time.Duration  `json:"scrape_interval" default:"120s"` // scrape interval
	Params         map[string]any `json:"params" default:"{}"`            // extra parameters
}

func LoadProviderFromConfig(ctx context.Context, cfg Config) (Provider, error) {
	factory, err := GetProvider(cfg.Type)
	if err != nil {
		return nil, err
	}
	return factory(ctx, cfg)
}
