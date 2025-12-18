package dhcp

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// Registry of all known sources
var AllSources = []LeaseSource{
	&DnsmasqSource{path: "/var/lib/misc/dnsmasq.leases"},
	&DnsmasqSource{path: "/tmp/dhcp.leases"},
	&DnsmasqSource{path: "/var/lib/dnsmasq/dnsmasq.leases"},
	&DnsmasqSource{path: "/tmp/dnsmasq.leases"},
	&ISCSource{path: "/var/lib/dhcp/dhcpd.leases"},
	&ISCSource{path: "/var/lib/dhcpd/dhcpd.leases"},
	// OpenWrt
	&DnsmasqSource{path: "/tmp/dhcp.leases"},
	// Merlin/Asus
	&DnsmasqSource{path: "/var/lib/misc/dnsmasq.leases"},
}

// === Dnsmasq ===
// Format: timestamp mac ip hostname clientid

type DnsmasqSource struct {
	path string
}

func (d *DnsmasqSource) Name() string { return "dnsmasq" }
func (d *DnsmasqSource) Path() string { return d.path }

func (d *DnsmasqSource) Detect() bool {
	_, err := os.Stat(d.path)
	return err == nil
}

func (d *DnsmasqSource) Parse() ([]Lease, error) {
	file, err := os.Open(d.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var leases []Lease
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		expires := time.Unix(0, 0)
		if ts, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
			expires = time.Unix(ts, 0)
		}

		hostname := ""
		if len(fields) >= 4 && fields[3] != "*" {
			hostname = fields[3]
		}

		leases = append(leases, Lease{
			MAC:      strings.ToUpper(fields[1]),
			IP:       fields[2],
			Hostname: hostname,
			Expires:  expires,
		})
	}

	return leases, scanner.Err()
}

// === ISC DHCP ===
// Format: lease { ... } blocks

type ISCSource struct {
	path string
}

func (i *ISCSource) Name() string { return "isc-dhcp" }
func (i *ISCSource) Path() string { return i.path }

func (i *ISCSource) Detect() bool {
	_, err := os.Stat(i.path)
	return err == nil
}

func (i *ISCSource) Parse() ([]Lease, error) {
	data, err := os.ReadFile(i.path)
	if err != nil {
		return nil, err
	}

	var leases []Lease
	var current Lease
	inLease := false

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "lease ") {
			inLease = true
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				current = Lease{IP: parts[1]}
			}
		} else if line == "}" && inLease {
			if current.IP != "" && current.MAC != "" {
				leases = append(leases, current)
			}
			current = Lease{}
			inLease = false
		} else if inLease {
			if strings.HasPrefix(line, "hardware ethernet ") {
				mac := strings.TrimSuffix(strings.TrimPrefix(line, "hardware ethernet "), ";")
				current.MAC = strings.ToUpper(mac)
			} else if strings.HasPrefix(line, "client-hostname ") {
				hostname := strings.TrimSuffix(strings.TrimPrefix(line, "client-hostname "), ";")
				current.Hostname = strings.Trim(hostname, "\"")
			} else if strings.HasPrefix(line, "ends ") {
				// Parse: ends 4 2024/01/15 12:00:00;
				parts := strings.Fields(strings.TrimSuffix(line, ";"))
				if len(parts) >= 4 {
					if t, err := time.Parse("2006/01/02 15:04:05", parts[2]+" "+parts[3]); err == nil {
						current.Expires = t
					}
				}
			}
		}
	}

	return leases, nil
}
