package discovery

import (
	"bufio"
	"bytes"
	"log/slog"
	"net"
	"os/exec"
	"strings"
	"time"

	"netdash/internal/device"
)

func StartARPWorker(store *device.Store) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		scanARPTable(store)
		<-ticker.C
	}
}

func scanARPTable(store *device.Store) {
	cmd := exec.Command("arp", "-an")
	out, err := cmd.Output()
	if err != nil {
		slog.Warn("arp: failed to run arp -an", "err", err)
		return
	}

	now := time.Now()
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		ip, ok := parseARPLine(scanner.Text())
		if !ok {
			continue
		}

		store.Update(ip, func(d *device.Device) {
			d.Online = true
			d.LastSeen = now
		})
		maybeResolveHostname(ip, store)
	}
}

func parseARPLine(line string) (string, bool) {
	start := strings.IndexByte(line, '(')
	end := strings.IndexByte(line, ')')
	if start == -1 || end == -1 || end <= start+1 {
		return "", false
	}

	ip := strings.TrimSpace(line[start+1 : end])
	if net.ParseIP(ip) == nil {
		return "", false
	}

	if strings.Contains(strings.ToLower(line), "incomplete") {
		return "", false
	}

	return ip, true
}
