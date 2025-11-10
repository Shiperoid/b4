// path: src/config/config.go
package config

import (
	_ "embed"

	"github.com/daniellavrushin/b4/log"
)

type UDPConfigV1 struct {
	Mode           string `json:"mode" bson:"mode"`
	FakeSeqLength  int    `json:"fake_seq_length" bson:"fake_seq_length"`
	FakeLen        int    `json:"fake_len" bson:"fake_len"`
	FakingStrategy string `json:"faking_strategy" bson:"faking_strategy"`
	DPortMin       int    `json:"dport_min" bson:"dport_min"`
	DPortMax       int    `json:"dport_max" bson:"dport_max"`
	FilterQUIC     string `json:"filter_quic" bson:"filter_quic"`
	FilterSTUN     bool   `json:"filter_stun" bson:"filter_stun"`
	ConnBytesLimit int    `json:"conn_bytes_limit" bson:"conn_bytes_limit"`
}

type ConfigV1 struct {
	ConfigPath     string  `json:"-" bson:"-"`
	QueueStartNum  int     `json:"queue_start_num" bson:"queue_start_num"`
	Mark           uint    `json:"mark" bson:"mark"`
	ConnBytesLimit int     `json:"conn_bytes_limit" bson:"conn_bytes_limit"`
	Logging        Logging `json:"logging" bson:"logging"`
	Threads        int     `json:"threads" bson:"threads"`
	Seg2Delay      int     `json:"seg2delay" bson:"seg2delay"`
	IPv4Enabled    bool    `json:"ipv4" bson:"ipv4"`
	IPv6Enabled    bool    `json:"ipv6" bson:"ipv6"`

	Domains       DomainsConfig       `json:"domains" bson:"domains"`
	Fragmentation FragmentationConfig `json:"fragmentation" bson:"fragmentation"`
	Faking        FakingConfig        `json:"faking" bson:"faking"`
	UDP           UDPConfigV1         `json:"udp" bson:"udp"`

	WebServer WebServerConfig `json:"web_server" bson:"web_server"`
	Tables    TablesConfig    `json:"tables" bson:"tables"`

	Checker CheckerConfig `json:"checker" bson:"checker"`
}
type DomainsConfig struct {
	GeoSitePath       string   `json:"geosite_path" bson:"geosite_path"`
	GeoIpPath         string   `json:"geoip_path" bson:"geoip_path"`
	SNIDomains        []string `json:"sni_domains" bson:"sni_domains"`
	GeoSiteCategories []string `json:"geosite_categories" bson:"geosite_categories"`
	GeoIpCategories   []string `json:"geoip_categories" bson:"geoip_categories"`
}

var DefaultConfigV1 = ConfigV1{
	ConfigPath:     "",
	QueueStartNum:  537,
	Mark:           1 << 15,
	Threads:        4,
	ConnBytesLimit: 19,
	Seg2Delay:      0,
	IPv4Enabled:    true,
	IPv6Enabled:    false,

	Domains: DomainsConfig{
		GeoSitePath:       "",
		GeoIpPath:         "",
		SNIDomains:        []string{},
		GeoSiteCategories: []string{},
		GeoIpCategories:   []string{},
	},

	Fragmentation: FragmentationConfig{
		Strategy:    "tcp",
		SNIReverse:  true,
		MiddleSNI:   true,
		SNIPosition: 1,
	},

	Faking: FakingConfig{
		SNI:           true,
		TTL:           8,
		SNISeqLength:  1,
		SNIType:       FakePayloadDefault,
		CustomPayload: "",
		Strategy:      "pastseq",
		SeqOffset:     10000,
	},

	UDP: UDPConfigV1{
		Mode:           "fake",
		FakeSeqLength:  6,
		FakeLen:        64,
		FakingStrategy: "none",
		DPortMin:       0,
		DPortMax:       0,
		FilterQUIC:     "disabled",
		FilterSTUN:     true,
		ConnBytesLimit: 8,
	},

	Tables: TablesConfig{
		MonitorInterval: 10,
		SkipSetup:       false,
	},

	WebServer: WebServerConfig{
		Port:      7000,
		IsEnabled: true,
	},

	Logging: Logging{
		Level:      log.LevelInfo,
		Instaflush: true,
		Syslog:     false,
	},

	Checker: CheckerConfig{
		TimeoutSeconds: 15,
		MaxConcurrent:  4,
		Domains:        []string{},
	},
}
