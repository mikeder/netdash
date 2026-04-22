package discovery

import (
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"time"

	"github.com/mdlayher/arp"

	"netdash/internal/device"
)

func StartActiveARPScanner(subnet string, store *device.Store) {
	iface, err := subnetInterface(subnet)
	if err != nil {
		slog.Warn("arp-scan: no interface found for subnet", "subnet", subnet, "err", err)
		return
	}

	slog.Info("arp-scan: starting active ARP scanner", "iface", iface.Name, "subnet", subnet+".0/24")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		if err := activeARPScan(iface, subnet, store); err != nil {
			slog.Warn("arp-scan: scan failed", "iface", iface.Name, "err", err)
		}
		<-ticker.C
	}
}

func activeARPScan(iface *net.Interface, subnet string, store *device.Store) error {
	client, err := arp.Dial(iface)
	if err != nil {
		return fmt.Errorf("dial (requires root/CAP_NET_RAW): %w", err)
	}
	defer client.Close()

	if err := client.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}

	for i := 1; i < 255; i++ {
		ip := net.ParseIP(fmt.Sprintf("%s.%d", subnet, i)).To4()
		if ip == nil {
			continue
		}
		addr, ok := netip.AddrFromSlice(ip)
		if !ok {
			continue
		}
		if err := client.Request(addr.Unmap()); err != nil {
			slog.Warn("arp-scan: request failed", "ip", addr, "err", err)
		}
	}

	now := time.Now()
	for {
		pkt, _, err := client.Read()
		if err != nil {
			break
		}
		if pkt.Operation != arp.OperationReply {
			continue
		}
		ip := pkt.SenderIP.String()
		if net.ParseIP(ip) == nil {
			continue
		}
		store.Update(ip, func(d *device.Device) {
			d.Online = true
			d.LastSeen = now
		})
		maybeResolveHostname(ip, store)
	}

	return nil
}

func subnetInterface(subnet string) (*net.Interface, error) {
	_, network, err := net.ParseCIDR(subnet + ".0/24")
	if err != nil {
		return nil, fmt.Errorf("parse subnet: %w", err)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && network.Contains(ip) {
				return &iface, nil
			}
		}
	}

	return nil, fmt.Errorf("no interface found for %s/24", subnet)
}
