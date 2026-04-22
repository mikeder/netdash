package discovery

import (
	"log/slog"
	"net"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"netdash/internal/device"
)

func StartMDNSWorker(store *device.Store) {
	go sendMDNSQueries()
	listenMDNS(store)
}

func listenMDNS(store *device.Store) {
	group := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	conn, err := net.ListenMulticastUDP("udp4", nil, group)
	if err != nil {
		slog.Warn("mdns: failed to join multicast group", "err", err)
		return
	}
	defer conn.Close()

	_ = conn.SetReadBuffer(64 * 1024)

	buf := make([]byte, 64*1024)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		now := time.Now()
		if src != nil && src.IP != nil {
			ip := src.IP.String()
			store.Update(ip, func(d *device.Device) {
				d.Online = true
				d.LastSeen = now
			})
			maybeResolveHostname(ip, store)
		}

		for _, ip := range extractIPsFromMDNS(buf[:n]) {
			store.Update(ip, func(d *device.Device) {
				d.Online = true
				d.LastSeen = now
			})
			maybeResolveHostname(ip, store)
		}
	}
}

func sendMDNSQueries() {
	target := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	conn, err := net.DialUDP("udp4", nil, target)
	if err != nil {
		slog.Warn("mdns: failed to dial multicast", "err", err)
		return
	}
	defer conn.Close()

	pkt, err := buildMDNSServicesQuery()
	if err != nil {
		slog.Warn("mdns: failed to build services query", "err", err)
		return
	}

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		if _, err := conn.Write(pkt); err != nil {
			slog.Warn("mdns: failed to send query", "err", err)
		}
		<-ticker.C
	}
}

func buildMDNSServicesQuery() ([]byte, error) {
	name, err := dnsmessage.NewName("_services._dns-sd._udp.local.")
	if err != nil {
		return nil, err
	}

	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{Response: false})
	b.EnableCompression()
	if err := b.StartQuestions(); err != nil {
		return nil, err
	}
	if err := b.Question(dnsmessage.Question{Name: name, Type: dnsmessage.TypePTR, Class: dnsmessage.ClassINET}); err != nil {
		return nil, err
	}

	return b.Finish()
}

func extractIPsFromMDNS(packet []byte) []string {
	var p dnsmessage.Parser
	_, err := p.Start(packet)
	if err != nil {
		return nil
	}

	for {
		_, err := p.Question()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			return nil
		}
	}

	ipSet := map[string]struct{}{}
	for {
		h, err := p.AnswerHeader()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			break
		}

		switch h.Type {
		case dnsmessage.TypeA:
			r, err := p.AResource()
			if err != nil {
				continue
			}
			ipSet[net.IP(r.A[:]).String()] = struct{}{}
		case dnsmessage.TypeAAAA:
			r, err := p.AAAAResource()
			if err != nil {
				continue
			}
			ipSet[net.IP(r.AAAA[:]).String()] = struct{}{}
		default:
			if err := p.SkipAnswer(); err != nil {
				continue
			}
		}
	}

	ipList := make([]string, 0, len(ipSet))
	for ip := range ipSet {
		if net.ParseIP(ip) == nil {
			continue
		}
		ipList = append(ipList, ip)
	}

	return ipList
}
