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
	// Try raw socket first; fall back to unprivileged ping socket
	// (udp4 requires net.ipv4.ping_group_range to include this process's GID).
	conn, udp, err := openICMPConn()
	if err != nil {
		return err
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
		var dst net.Addr
		if udp {
			dst = &net.UDPAddr{IP: ip}
		} else {
			dst = &net.IPAddr{IP: ip}
		}
		if _, err := conn.WriteTo(b, dst); err != nil {
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
		var ip string
		if udp {
			ip = peer.(*net.UDPAddr).IP.String()
		} else {
			ip = peer.(*net.IPAddr).IP.String()
		}
		store.Update(ip, func(d *device.Device) {
			d.Online = true
			d.LastSeen = now
		})
		maybeResolveHostname(ip, store)
	}

	return nil
}

// openICMPConn tries a raw ICMP socket first (needs CAP_NET_RAW), then falls
// back to the unprivileged ping socket (needs net.ipv4.ping_group_range to
// include this process's GID, or --sysctl net.ipv4.ping_group_range=0 2147483647).
// Returns the connection and whether it is a UDP-mode socket.
func openICMPConn() (*icmp.PacketConn, bool, error) {
	if conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0"); err == nil {
		return conn, false, nil
	}
	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return nil, false, fmt.Errorf("open ICMP socket (need CAP_NET_RAW or net.ipv4.ping_group_range): %w", err)
	}
	return conn, true, nil
}
