package discovery

import (
	"context"
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
	dnsResolver         = net.DefaultResolver
)

const hostnameLookupCooldown = 5 * time.Minute

// SetDNSServer directs reverse lookups to addr (e.g. "192.168.1.1") instead
// of the system resolver. Useful inside Docker where the default resolver
// won't answer PTR queries for LAN addresses.
func SetDNSServer(addr string) {
	dnsResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "udp", net.JoinHostPort(addr, "53"))
		},
	}
	slog.Info("hostname: using DNS server", "addr", addr)
}

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

		names, err := dnsResolver.LookupAddr(context.Background(), ip)
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
