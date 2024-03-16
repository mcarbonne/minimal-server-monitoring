package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/alert"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/config"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

func usage() {
	fmt.Println("Usage: " + os.Args[0] + " config.json")
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

	scrapeResultChan := make(chan provider.ScrapeResult, 5)
	notifyChan := make(chan notifier.Message, 5)

	go func() {
		<-signalChan
		logging.Info("Exiting...")
		storage.Sync(false)
		os.Exit(1)
	}()

	// alert center (send notifications)
	go func() {
		alert.AlertCenter(cfg.Alert, scrapeResultChan, notifyChan)
	}()

	go func() {
		notifier.LoadAndRunNotifiers(cfg.MachineName, cfg.Notifiers, notifyChan)
	}()

	// Start metric scraping
	scraping.ScheduleScraping(cfg.Providers, storage, scrapeResultChan)
}
