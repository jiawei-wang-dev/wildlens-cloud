package main

import (
	"context"
	"log"

	appconfig "github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/config"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/handler"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/router"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/service"
)

func main() {
	ctx := context.Background()

	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("load application configuration: %v", err)
	}

	repo, err := buildMediaRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("create media repository: %v", err)
	}

	urlSigner, err := buildMediaURLSigner(ctx, cfg)
	if err != nil {
		log.Fatalf("create media URL signer: %v", err)
	}

	queryService := service.NewQueryServiceWithURLSigner(
		repo,
		urlSigner,
	)
	queryHandler := handler.NewQueryHandler(queryService)

	engine := router.New(queryHandler)

	log.Printf(
		"WildLens query API is running at http://localhost:8080 using %s repository",
		cfg.Repository,
	)

	if err := engine.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
