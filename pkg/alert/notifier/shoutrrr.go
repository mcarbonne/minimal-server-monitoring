package notifier

import (
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
)

type Shoutrrr struct {
	router *router.ServiceRouter
}

func NewShoutrrr(url string) (Notifier, error) {
	router, err := shoutrrr.CreateSender(url)
	if err != nil {
		return nil, err
	}
	return &Shoutrrr{router: router}, nil
}

func (shoutrrr *Shoutrrr) Send(message Message) error {
	params := types.Params{"title": message.Title}
	return shoutrrr.router.Send(message.Message, &params)[0]
}
