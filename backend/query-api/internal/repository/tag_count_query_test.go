package repository

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

func TestMemoryRepositoryFindByTagCountsMinimums(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID: "checksum-image-001",
			TagCounts: map[string]int{
				"koala":  3,
				"magpie": 1,
			},
		},
	})

	files, err := repo.FindByTagCounts(
		context.Background(),
		map[string]int{"koala": 3},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	files, err = repo.FindByTagCounts(
		context.Background(),
		map[string]int{"koala": 4},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestMemoryRepositoryFindByTagCountsUsesAND(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID: "checksum-image-001",
			TagCounts: map[string]int{
				"koala":  3,
				"magpie": 1,
			},
		},
	})

	files, err := repo.FindByTagCounts(
		context.Background(),
		map[string]int{
			"koala":  3,
			"magpie": 1,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	files, err = repo.FindByTagCounts(
		context.Background(),
		map[string]int{
			"koala":  3,
			"magpie": 2,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestDynamoDBRepositoryFindByTagCountsMinimums(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID: "checksum-image-001",
		TagCounts: map[string]int{
			"koala":  3,
			"magpie": 1,
		},
	})

	client := &fakeDynamoDBClient{
		pages: []*dynamodb.ScanOutput{
			{
				Items: []map[string]types.AttributeValue{
					item,
				},
			},
		},
	}

	repo := NewDynamoDBRepository(client, "test-table")

	files, err := repo.FindByTagCounts(
		context.Background(),
		map[string]int{"koala": 3},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestDynamoDBRepositoryFindByTagCountsUsesANDMinimums(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID: "checksum-image-001",
		TagCounts: map[string]int{
			"koala":  3,
			"magpie": 1,
		},
	})

	client := &fakeDynamoDBClient{
		pages: []*dynamodb.ScanOutput{
			{
				Items: []map[string]types.AttributeValue{
					item,
				},
			},
		},
	}

	repo := NewDynamoDBRepository(client, "test-table")

	files, err := repo.FindByTagCounts(
		context.Background(),
		map[string]int{
			"koala":  3,
			"magpie": 2,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}
