package config

import (
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/alert"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper"
)

type Config struct {
	MachineName string                     `json:"machine_name" default:""`
	Notifiers   map[string]notifier.Config `json:"notifiers"`
	Alert       alert.Config               `json:"alert" default:"{}"`
	Scrapers    map[string]provider.Config `json:"scrapers"`
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
		logging.Fatal("%v", err.Error())
	}

	var rawYaml map[string]interface{}

	configFile := os.Expand(string(configFileBytes), strictGetEnv)

	yamlParser := yaml.NewDecoder(strings.NewReader(configFile))
	err = yamlParser.Decode(&rawYaml)
	if err != nil {
		logging.Fatal("%v", err.Error())
	}

	config, err := configmapper.MapOnStruct[Config](rawYaml)
	if err != nil {
		logging.Fatal("Error loading config: %v", err)
	}

	if config.MachineName == "" {
		config.MachineName, _ = os.Hostname()
	}

	return config
}
