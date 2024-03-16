package provider

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

type Provider interface {
	Update(result *ScrapeResult, storage storage.Storager)
}
