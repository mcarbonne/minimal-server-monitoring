package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/alert/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

type Config struct {
	Notifiers []notifier.Config `json:"notifiers"`
	CachePath string            `json:"cache"`
}

func strictGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		logging.Fatal("Missing or empty environment variable '%v'", key)
	}
	return value
}

func LoadConfiguration(configPath string) Config {
	var config Config
	configFileBytes, err := os.ReadFile(configPath)

	if err != nil {
		logging.Fatal(err.Error())
	}

	configFile := os.Expand(string(configFileBytes), strictGetEnv)

	jsonParser := json.NewDecoder(strings.NewReader(configFile))
	jsonParser.DisallowUnknownFields()
	err = jsonParser.Decode(&config)
	if err != nil {
		logging.Fatal(err.Error())
	}
	return config
}
