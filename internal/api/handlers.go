package api

import (
	"encoding/json"
	"net/http"

	"netdash/internal/alerts"
	"netdash/internal/device"
)

func RegisterRoutes(mux *http.ServeMux, store *device.Store, alerts *alerts.Manager) {

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(store.All())
	})
}
