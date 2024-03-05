package metric

import "time"

type Policy struct {
	refreshInterval time.Duration
}

func MakePolicy(refreshInterval time.Duration) Policy {
	return Policy{refreshInterval: refreshInterval}
}
