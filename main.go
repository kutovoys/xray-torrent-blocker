package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tblocker/config"
	"tblocker/storage"
	"tblocker/utils"
)

var Version string

func main() {
	initConfig()

	log.Printf("XRay torrent-blocker: %s", Version)
	log.Printf("Service started on %s", config.Hostname)

	utils.StartLogMonitor()
}

func initConfig() {
	var configPath string
	var showVersion bool

	flag.StringVar(&configPath, "c", "", "Path to the configuration file")
	flag.BoolVar(&showVersion, "v", false, "Display version")
	flag.Parse()

	if showVersion {
		fmt.Printf("XRay torrent-blocker: %s\n", Version)
		os.Exit(0)
	}

	if configPath == "" {
		ex, err := os.Executable()
		if err != nil {
			log.Fatalf("Error getting executable path: %v", err)
		}
		configPath = filepath.Join(filepath.Dir(ex), "config.yaml")
	}

	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	store, err := storage.NewIPStorage(config.StorageDir, utils.UnblockIPAfterDelay)
	if err != nil {
		log.Fatalf("Failed to initialize IP storage: %v", err)
	}
	utils.SetIPStorage(store)

	utils.ScheduleBlockedIPsUpdate()
}
