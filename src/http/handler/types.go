package handler

import (
	"github.com/daniellavrushin/b4/config"
)

type API struct {
	cfg *config.Config
}

type GeositeResponse struct {
	Tags []string `json:"tags"`
}
