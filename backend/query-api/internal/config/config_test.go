package config

import "testing"

func TestLoadDefaultsToMemoryRepository(t *testing.T) {
	t.Setenv("APP_REPOSITORY", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("DYNAMODB_TABLE_NAME", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Repository != RepositoryMemory {
		t.Fatalf(
			"expected repository %q, got %q",
			RepositoryMemory,
			cfg.Repository,
		)
	}

	if cfg.AWSRegion != "us-east-1" {
		t.Fatalf(
			"expected region us-east-1, got %q",
			cfg.AWSRegion,
		)
	}
}

func TestLoadRejectsDynamoDBWithoutTableName(t *testing.T) {
	t.Setenv("APP_REPOSITORY", RepositoryDynamoDB)
	t.Setenv("DYNAMODB_TABLE_NAME", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestLoadAcceptsDynamoDBConfiguration(t *testing.T) {
	t.Setenv("APP_REPOSITORY", RepositoryDynamoDB)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv(
		"DYNAMODB_TABLE_NAME",
		"fit5225-wildlife-media-metadata",
	)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Repository != RepositoryDynamoDB {
		t.Fatalf(
			"expected repository %q, got %q",
			RepositoryDynamoDB,
			cfg.Repository,
		)
	}

	if cfg.DynamoDBTableName != "fit5225-wildlife-media-metadata" {
		t.Fatalf(
			"unexpected table name: %q",
			cfg.DynamoDBTableName,
		)
	}
}

func TestLoadRejectsUnsupportedRepository(t *testing.T) {
	t.Setenv("APP_REPOSITORY", "mysql")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}
