package main

import (
	appconfig "github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/config"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/inference"
)

func buildImageInferenceClient(
	cfg appconfig.AppConfig,
) inference.ImageClient {
	if cfg.TempQueryInferURL == "" {
		return inference.NewNoopClient()
	}

	return inference.NewHTTPClient(cfg.TempQueryInferURL, nil)
}
