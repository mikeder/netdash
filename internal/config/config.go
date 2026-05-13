package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"netdash/internal/device"
)

type Config struct {
	Subnet    string
	Interface string
	ScanPorts []int
	DNSServer string
}

func Load() Config {
	subnet := getEnvOr("NETDASH_SUBNET", "192.168.1")
	return Config{
		Subnet:    subnet,
		Interface: os.Getenv("NETDASH_INTERFACE"),
		ScanPorts: parseScanPorts(os.Getenv("NETDASH_SCAN_PORTS"), []int{22, 53, 80, 110, 139, 143, 443, 445, 587, 993, 995, 3306, 3389, 5432, 8080}),
		DNSServer: getEnvOr("NETDASH_DNS", subnet+".1"),
	}
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// labelEntry is one record in devices.json. Key by "ip" or "mac".
type labelEntry struct {
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
	Label  string `json:"label"`
	Vendor string `json:"vendor"`
}

// LoadLabels reads devices.json and pre-seeds per-device labels/vendor overrides.
// Entries are matched by IP or MAC address.
func LoadLabels(store *device.Store, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("labels: failed to read file", "path", path, "err", err)
		}
		return
	}

	var entries []labelEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		slog.Warn("labels: failed to parse file", "path", path, "err", err)
		return
	}

	for _, e := range entries {
		e := e
		if e.IP != "" {
			store.Update(e.IP, func(d *device.Device) {
				if e.Label != "" {
					d.Label = e.Label
				}
				if e.Vendor != "" {
					d.Vendor = e.Vendor
				}
				if e.MAC != "" {
					d.MAC = e.MAC
				}
			})
		}
	}
	slog.Info("labels: loaded", "path", path, "count", len(entries))
}

func parseScanPorts(raw string, fallback []int) []int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}

	parts := strings.Split(raw, ",")
	ports := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		port, err := strconv.Atoi(p)
		if err != nil || port < 1 || port > 65535 {
			continue
		}

		ports = append(ports, port)
	}

	if len(ports) == 0 {
		return fallback
	}

	return ports
}
