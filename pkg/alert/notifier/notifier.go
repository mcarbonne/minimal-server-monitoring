package notifier

import "github.com/mcarbonne/minimal-server-monitoring/pkg/logging"

type Notifier interface {
	Send(message Message) error
}

func LoadNotifierFromConfig(cfg Config) Notifier {
	var n Notifier
	var err error
	switch cfg.Type {
	case "shoutrrr":
		url, found := cfg.Params["url"]
		if !found {
			logging.Fatal("missing url in %+v", cfg)
		}
		n, err = NewShoutrrr(url)
	case "log":
		n = NewLog()
	default:
		logging.Fatal("Illegal notifier type: %v", cfg.Type)
	}
	if err != nil || n == nil {
		logging.Fatal("Unable to setup notifier for %v: %v", cfg.Type, err.Error())
	}
	return n
}
