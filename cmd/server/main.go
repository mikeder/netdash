package main

import (
	"log"
	"net/http"

	"netdash/internal/alerts"
	"netdash/internal/api"
	"netdash/internal/config"
	"netdash/internal/device"
	"netdash/internal/discovery"
	"netdash/internal/storage"
)

func main() {
	cfg := config.Load()

	store := device.NewStore()
	db := storage.InitDB("netdash.db")
	alertMgr := alerts.NewManager(db)

	store.SetAlertChannel(alertMgr.Channel())
	store.SetDB(db)

	config.LoadLabels(store, "devices.json")

	go discovery.StartScanner(cfg.Subnet, cfg.ScanPorts, store)
	go discovery.StartARPWorker(store)
	go discovery.StartMDNSWorker(store)

	go alertMgr.Run()

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, store, alertMgr)

	log.Println("Running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
