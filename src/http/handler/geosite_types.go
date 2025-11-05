package handler

import "github.com/daniellavrushin/b4/config"

// Response types for API endpoints
type GeositeResponse struct {
	Tags []string `json:"tags"`
}

// ConfigResponse wraps the config with additional metadata
type ConfigResponse struct {
	*config.Config
	DomainStats DomainStatistics `json:"domain_stats"`
}

// DomainStatistics provides overview of domain configuration
type DomainStatistics struct {
	ManualDomains     int            `json:"manual_domains"`
	GeositeDomains    int            `json:"geosite_domains"`
	TotalDomains      int            `json:"total_domains"`
	CategoryBreakdown map[string]int `json:"category_breakdown,omitempty"`
	GeositeAvailable  bool           `json:"geosite_available"`
}

// CategoryPreviewResponse for previewing category contents
type CategoryPreviewResponse struct {
	Category     string   `json:"category"`
	TotalDomains int      `json:"total_domains"`
	PreviewCount int      `json:"preview_count"`
	Preview      []string `json:"preview"`
}
