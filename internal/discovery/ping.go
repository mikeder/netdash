package discovery

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"netdash/internal/device"
)

func StartPingSweep(subnet string, store *device.Store) {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	for {
		if err := pingSweep(subnet, store); err != nil {
			slog.Warn("ping: sweep failed", "err", err)
		}
		<-ticker.C
	}
}

func pingSweep(subnet string, store *device.Store) error {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("open ICMP socket (requires root/CAP_NET_RAW): %w", err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}

	pid := os.Getpid() & 0xffff

	for i := 1; i < 255; i++ {
		ip := net.ParseIP(fmt.Sprintf("%s.%d", subnet, i))
		if ip == nil {
			continue
		}
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{ID: pid, Seq: i, Data: []byte("netdash")},
		}
		b, err := msg.Marshal(nil)
		if err != nil {
			continue
		}
		if _, err := conn.WriteTo(b, &net.IPAddr{IP: ip}); err != nil {
			slog.Warn("ping: send failed", "ip", ip, "err", err)
		}
	}

	buf := make([]byte, 1500)
	now := time.Now()

	for {
		n, peer, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}
		msg, err := icmp.ParseMessage(1, buf[:n])
		if err != nil || msg.Type != ipv4.ICMPTypeEchoReply {
			continue
		}
		ip := peer.(*net.IPAddr).IP.String()
		store.Update(ip, func(d *device.Device) {
			d.Online = true
			d.LastSeen = now
		})
		maybeResolveHostname(ip, store)
	}

	return nil
}
