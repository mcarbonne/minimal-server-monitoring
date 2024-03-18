package notifier

type params map[string]any

type Config struct {
	Type   string `json:"type"`
	Params params `json:"params" default:"{}"`
}
