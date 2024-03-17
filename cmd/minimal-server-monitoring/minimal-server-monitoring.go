package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/alert"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/config"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
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

	scrapeResultChan := make(chan any, 5)
	notifyChan := make(chan notifier.Message, 5)

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(2)
	// alert center
	alert.AlertCenter(ctx, cfg.Alert, scrapeResultChan, notifyChan)

	// notifier
	go func() {
		defer wg.Done()
		notifier.LoadAndRunNotifiers(ctx, cfg.MachineName, cfg.Notifiers, notifyChan)
	}()

	// Start metric scraping
	go func() {
		defer wg.Done()
		scraping.ScheduleScraping(ctx, cfg.Scrapers, storage, scrapeResultChan)
	}()

	<-signalChan
	logging.Info("Exiting...")
	storage.Sync(false)
	cancel()
	logging.Info("Waiting for jobs to finish")
	if utils.WaitTimeout(&wg, 5*time.Second) {
		logging.Error("Failed to exit properly")
	}

}
