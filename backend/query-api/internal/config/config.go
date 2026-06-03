package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	// RepositoryMemory uses temporary in-memory data.
	RepositoryMemory = "memory"

	// RepositoryDynamoDB uses the AWS DynamoDB table.
	RepositoryDynamoDB = "dynamodb"
)

// AppConfig contains runtime configuration for the query API.
type AppConfig struct {
	Repository        string
	AWSRegion         string
	DynamoDBTableName string
}

// Load reads application configuration from environment variables.
func Load() (AppConfig, error) {
	cfg := AppConfig{
		Repository: envOrDefault(
			"APP_REPOSITORY",
			RepositoryMemory,
		),
		AWSRegion: envOrDefault(
			"AWS_REGION",
			"us-east-1",
		),
		DynamoDBTableName: strings.TrimSpace(
			os.Getenv("DYNAMODB_TABLE_NAME"),
		),
	}

	switch cfg.Repository {
	case RepositoryMemory:
		return cfg, nil

	case RepositoryDynamoDB:
		if cfg.DynamoDBTableName == "" {
			return AppConfig{}, fmt.Errorf(
				"DYNAMODB_TABLE_NAME is required when APP_REPOSITORY=%s",
				RepositoryDynamoDB,
			)
		}

		return cfg, nil

	default:
		return AppConfig{}, fmt.Errorf(
			"unsupported APP_REPOSITORY value: %q",
			cfg.Repository,
		)
	}
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return fallback
	}

	return value
}
