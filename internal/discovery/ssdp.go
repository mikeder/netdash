package discovery

import (
	"log/slog"
	"net"
	"time"

	"netdash/internal/device"
)

const (
	ssdpGroup = "239.255.255.250:1900"
	ssdpQuery = "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nMAN: \"ssdp:discover\"\r\nMX: 3\r\nST: ssdp:all\r\n\r\n"
)

func StartSSDPWorker(store *device.Store) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		scanSSDP(store)
		<-ticker.C
	}
}

func scanSSDP(store *device.Store) {
	dest, err := net.ResolveUDPAddr("udp4", ssdpGroup)
	if err != nil {
		slog.Warn("ssdp: resolve address failed", "err", err)
		return
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		slog.Warn("ssdp: listen failed", "err", err)
		return
	}
	defer conn.Close()

	if _, err := conn.WriteTo([]byte(ssdpQuery), dest); err != nil {
		slog.Warn("ssdp: M-SEARCH failed", "err", err)
		return
	}

	if err := conn.SetReadDeadline(time.Now().Add(4 * time.Second)); err != nil {
		slog.Warn("ssdp: set deadline failed", "err", err)
		return
	}

	buf := make([]byte, 2048)
	now := time.Now()

	for {
		_, src, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}
		ip := src.(*net.UDPAddr).IP.String()
		if net.ParseIP(ip) == nil {
			continue
		}
		store.Update(ip, func(d *device.Device) {
			d.Online = true
			d.LastSeen = now
		})
		maybeResolveHostname(ip, store)
	}
}
