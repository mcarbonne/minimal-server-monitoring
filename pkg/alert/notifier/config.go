package notifier

type Config struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}
