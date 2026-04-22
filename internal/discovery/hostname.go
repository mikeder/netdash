package discovery

import (
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"netdash/internal/device"
)

var (
	hostnameMu          sync.Mutex
	hostnameLastAttempt = make(map[string]time.Time)
	hostnameInFlight    = make(map[string]bool)
)

const hostnameLookupCooldown = 5 * time.Minute

func maybeResolveHostname(ip string, store *device.Store) {
	if net.ParseIP(ip) == nil {
		return
	}

	now := time.Now()
	hostnameMu.Lock()
	if hostnameInFlight[ip] {
		hostnameMu.Unlock()
		return
	}
	if last, ok := hostnameLastAttempt[ip]; ok && now.Sub(last) < hostnameLookupCooldown {
		hostnameMu.Unlock()
		return
	}

	hostnameInFlight[ip] = true
	hostnameLastAttempt[ip] = now
	hostnameMu.Unlock()

	go func(ip string) {
		defer func() {
			hostnameMu.Lock()
			delete(hostnameInFlight, ip)
			hostnameMu.Unlock()
		}()

		names, err := net.LookupAddr(ip)
		if err != nil {
			slog.Warn("hostname: reverse lookup failed", "ip", ip, "err", err)
			return
		}

		hostname := normalizeHostname(names)
		if hostname == "" {
			return
		}

		store.Update(ip, func(d *device.Device) {
			d.Hostname = hostname
		})
	}(ip)
}

func normalizeHostname(names []string) string {
	for _, name := range names {
		host := strings.TrimSuffix(strings.TrimSpace(name), ".")
		if host != "" {
			return host
		}
	}

	return ""
}
