package main

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	appconfig "github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/config"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/repository"
)

// buildMediaRepository selects the data source based on runtime configuration.
func buildMediaRepository(
	ctx context.Context,
	cfg appconfig.AppConfig,
) (repository.MediaRepository, error) {
	switch cfg.Repository {
	case appconfig.RepositoryMemory:
		return repository.NewMemoryRepository(seedMediaFiles()), nil

	case appconfig.RepositoryDynamoDB:
		awsCfg, err := awsconfig.LoadDefaultConfig(
			ctx,
			awsconfig.WithRegion(cfg.AWSRegion),
		)
		if err != nil {
			return nil, fmt.Errorf("load AWS configuration: %w", err)
		}

		client := dynamodb.NewFromConfig(awsCfg)

		return repository.NewDynamoDBRepository(
			client,
			cfg.DynamoDBTableName,
		), nil

	default:
		return nil, fmt.Errorf(
			"unsupported repository mode: %q",
			cfg.Repository,
		)
	}
}

// seedMediaFiles returns temporary data for local memory mode.
func seedMediaFiles() []model.MediaFile {
	return []model.MediaFile{
		{
			FileID:              "checksum-image-001",
			OriginalFilename:    "koala.jpg",
			FileType:            "image",
			MimeType:            "image/jpeg",
			ChecksumSHA256:      "checksum-image-001",
			StorageProvider:     "aws",
			Bucket:              "wildlens-media",
			ObjectPath:          "media/originals/koala.jpg",
			ThumbnailObjectPath: "media/thumbnails/koala.jpg",
			FileURL:             "s3://wildlens-media/media/originals/koala.jpg",
			ThumbnailURL:        "s3://wildlens-media/media/thumbnails/koala.jpg",
			Tags:                []string{"koala", "magpie"},
			TagCounts: map[string]int{
				"koala":  3,
				"magpie": 1,
			},
			Status: "ready",
		},
		{
			FileID:           "checksum-video-001",
			OriginalFilename: "wombat.mp4",
			FileType:         "video",
			MimeType:         "video/mp4",
			ChecksumSHA256:   "checksum-video-001",
			StorageProvider:  "aws",
			Bucket:           "wildlens-media",
			ObjectPath:       "media/originals/wombat.mp4",
			FileURL:          "s3://wildlens-media/media/originals/wombat.mp4",
			Tags:             []string{"wombat"},
			TagCounts: map[string]int{
				"wombat": 2,
			},
			Status: "ready",
		},
	}
}
