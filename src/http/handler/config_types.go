package handler

import "github.com/daniellavrushin/b4/config"

// ConfigUpdateRequest for handling config updates
type ConfigUpdateRequest struct {
	*config.Config
	// Additional fields for UI state if needed
	PreserveManuaDomains bool `json:"preserve_manual_domains,omitempty"`
}

// ConfigUpdateResponse for config update results
type ConfigUpdateResponse struct {
	Success     bool             `json:"success"`
	Message     string           `json:"message"`
	DomainStats DomainStatistics `json:"domain_stats"`
	Warnings    []string         `json:"warnings,omitempty"`
}
