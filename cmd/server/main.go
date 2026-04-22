package main

import (
	"log/slog"
	"net/http"
	"os"

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
	go discovery.StartActiveARPScanner(cfg.Subnet, store)
	go discovery.StartPingSweep(cfg.Subnet, store)
	go discovery.StartMDNSWorker(store)
	go discovery.StartSSDPWorker(store)

	go alertMgr.Run()

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, store, alertMgr)

	slog.Info("Running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("server exited", "err", err)
		os.Exit(1)
	}
}
