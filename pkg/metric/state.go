package metric

//go:generate stringer -type State -linecomment
type State uint

const (
	StateUndefined State = iota // Undefined
	StateHealthy                // Healthy
	StateUnhealthy              // Unhealthy
)
