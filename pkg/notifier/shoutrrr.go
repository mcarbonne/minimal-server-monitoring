package notifier

import (
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper"
)

type Shoutrrr struct {
	Url    string `json:"url"`
	router *router.ServiceRouter
}

func NewShoutrrr(params params) (Notifier, error) {
	notifier, err := configmapper.MapOnStruct[Shoutrrr](params)
	if err != nil {
		return nil, err
	}
	notifier.router, err = shoutrrr.CreateSender(notifier.Url)
	if err != nil {
		return nil, err
	}
	return &notifier, nil
}

func (shoutrrr *Shoutrrr) Send(message Message) error {
	params := types.Params{"title": message.Title}
	return shoutrrr.router.Send(message.Message, &params)[0]
}

func init() {
	RegisterNotifier("shoutrrr", func(cfg Config) (Notifier, error) {
		return NewShoutrrr(cfg.Params)
	})
}
