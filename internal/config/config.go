package config

import (
	"os"
	"strconv"
	"strings"

	"netdash/internal/device"
)

type Config struct {
	Subnet    string
	ScanPorts []int
}

func Load() Config {
	return Config{
		Subnet:    "192.168.1",
		ScanPorts: parseScanPorts(os.Getenv("NETDASH_SCAN_PORTS"), []int{22, 53, 80, 110, 139, 143, 443, 445, 587, 993, 995, 3306, 3389, 5432, 8080}),
	}
}

func LoadLabels(store *device.Store, path string) {}

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
