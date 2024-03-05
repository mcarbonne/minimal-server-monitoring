package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/alert"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/alert/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/config"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/metric"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/metric/providers"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

func usage() {
	fmt.Println("Usage: " + os.Args[0] + " config.json")
}

func eventToMessage(evt metric.Event) notifier.Message {
	return notifier.MakeMessage(evt.Topic, fmt.Sprintf("%v", evt.EventType), evt.Description)
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	configPath := os.Args[1]
	cfg := config.LoadConfiguration(configPath)

	storage := storage.NewJSONStorage(cfg.CachePath)
	storage.Sync(true) // Test if storage can be synced

	messageChannel := make(chan notifier.Message, 5)
	eventChannel := make(chan metric.Event, 5)

	go func() {
		<-signalChan
		logging.Info("Exiting...")
		storage.Sync(false)
		os.Exit(1)
	}()

	// alert center (send notifications)
	go func() {
		alert.AlertCenter(cfg.Notifiers, messageChannel)
	}()

	// transform events to notifications
	go func() {
		for evt := range eventChannel {
			messageChannel <- eventToMessage(evt)
		}
	}()

	// Start metric gathering
	var metricsToTest []*metric.Metric
	metricsToTest = append(metricsToTest, metric.MakeMetric("docker", providers.NewDockerProvider(), metric.MakePolicy(time.Second*5)))
	metric.ScheduleMetrics(metricsToTest, storage, eventChannel)
}
