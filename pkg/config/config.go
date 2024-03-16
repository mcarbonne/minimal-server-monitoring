package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/alert"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
)

type Config struct {
	MachineName string                     `json:"machine_name" default:""`
	Notifiers   []notifier.Config          `json:"notifiers"`
	Alert       alert.Config               `json:"alert"`
	Providers   map[string]provider.Config `json:"providers"`
	CachePath   string                     `json:"cache"`
}

func strictGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		logging.Fatal("Missing or empty environment variable '%v'", key)
	}
	return value
}

func LoadConfiguration(configPath string) Config {
	configFileBytes, err := os.ReadFile(configPath)

	if err != nil {
		logging.Fatal(err.Error())
	}

	var rawJson map[string]interface{}

	configFile := os.Expand(string(configFileBytes), strictGetEnv)

	jsonParser := json.NewDecoder(strings.NewReader(configFile))
	err = jsonParser.Decode(&rawJson)
	if err != nil {
		logging.Fatal(err.Error())
	}

	config, err := utils.MapOnStruct[Config](rawJson)
	if err != nil {
		logging.Fatal("Error loading config: %v", err)
	}

	if config.MachineName == "" {
		config.MachineName, _ = os.Hostname()
	}

	return config
}
