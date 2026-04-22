package discovery

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"netdash/internal/device"
)

var defaultScanPorts = []int{22, 53, 80, 110, 139, 143, 443, 445, 587, 993, 995, 3306, 3389, 5432, 8080}

var knownTCPServices = map[int]string{
	22:   "ssh",
	53:   "dns",
	80:   "http",
	110:  "pop3",
	139:  "netbios-ssn",
	143:  "imap",
	443:  "https",
	445:  "microsoft-ds",
	587:  "smtp-submission",
	993:  "imaps",
	995:  "pop3s",
	3306: "mysql",
	3389: "rdp",
	5432: "postgresql",
	8080: "http-alt",
}

func StartScanner(subnet string, ports []int, store *device.Store) {
	if len(ports) == 0 {
		ports = defaultScanPorts
	}

	for {
		for i := 1; i < 255; i++ {
			ip := fmt.Sprintf("%s.%d", subnet, i)

			go func(ip string) {
				openPorts := probeOpenPorts(ip, ports, 300*time.Millisecond)
				online := len(openPorts) > 0
				now := time.Now()

				store.Update(ip, func(d *device.Device) {
					if online {
						d.Online = true
						d.LastSeen = now
						d.Services = servicesFromPorts(openPorts)
					} else if time.Since(d.LastSeen) > 60*time.Second {
						d.Online = false
						d.Services = nil
					}
				})

				if online {
					maybeResolveHostname(ip, store)
				}
			}(ip)
		}
		time.Sleep(10 * time.Second)
	}
}

func probeOpenPorts(ip string, ports []int, timeout time.Duration) []int {
	open := make([]int, 0, len(ports))

	for _, port := range ports {
		addr := net.JoinHostPort(ip, strconv.Itoa(port))
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			continue
		}
		_ = conn.Close()
		open = append(open, port)
	}

	return open
}

func servicesFromPorts(ports []int) []device.Service {
	services := make([]device.Service, 0, len(ports))
	for _, port := range ports {
		name, ok := knownTCPServices[port]
		if !ok {
			name = "tcp"
		}

		services = append(services, device.Service{Port: port, Name: name})
	}

	return services
}
