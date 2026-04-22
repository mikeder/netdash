package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"netdash/internal/alerts"
	"netdash/internal/device"
)

func RegisterRoutes(mux *http.ServeMux, store *device.Store, alerts *alerts.Manager) {
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(store.All())
	})

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		send := func() {
			data, err := json.Marshal(store.All())
			if err != nil {
				slog.Warn("sse: failed to marshal devices", "err", err)
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		ch := store.Subscribe()
		defer store.Unsubscribe(ch)

		send()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ch:
				send()
			}
		}
	})
}
