package main

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appconfig "github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/config"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/storage"
)

// buildMediaURLSigner selects the signer used for media display URLs.
func buildMediaURLSigner(
	ctx context.Context,
	cfg appconfig.AppConfig,
) (storage.MediaURLSigner, error) {
	switch cfg.Repository {
	case appconfig.RepositoryMemory:
		return storage.NewStaticURLSigner(
			"https://local.wildlens.test",
		), nil

	case appconfig.RepositoryDynamoDB:
		awsCfg, err := awsconfig.LoadDefaultConfig(
			ctx,
			awsconfig.WithRegion(cfg.AWSRegion),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"load AWS configuration for S3 signer: %w",
				err,
			)
		}

		client := s3.NewFromConfig(awsCfg)
		presignClient := s3.NewPresignClient(client)

		return storage.NewS3URLSigner(
			presignClient,
			storage.DefaultPresignExpiry,
		), nil

	default:
		return nil, fmt.Errorf(
			"unsupported repository mode: %q",
			cfg.Repository,
		)
	}
}
