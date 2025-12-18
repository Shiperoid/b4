package dhcp

import "time"

type Lease struct {
	MAC      string
	IP       string
	Hostname string
	Expires  time.Time
}

type LeaseSource interface {
	Name() string
	Detect() bool
	Parse() ([]Lease, error)
	Path() string
}

type LeaseUpdateCallback func(ipToMAC map[string]string)
