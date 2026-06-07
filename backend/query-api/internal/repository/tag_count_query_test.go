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

func TestMemoryRepositoryFindByTagCountsNormalisesHistoricalKeys(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID: "checksum-image-001",
			TagCounts: map[string]int{
				"Alectura_lathami": 1,
				"magpie":           2,
			},
		},
		{
			FileID: "checksum-image-002",
			TagCounts: map[string]int{
				"alectura_lathami": 1,
			},
		},
	})

	files, err := repo.FindByTagCounts(
		context.Background(),
		map[string]int{"alectura_lathami": 1},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	files, err = repo.FindByTagCounts(
		context.Background(),
		map[string]int{"Alectura_lathami": 1},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	files, err = repo.FindByTagCounts(
		context.Background(),
		map[string]int{
			"alectura_lathami": 1,
			"magpie":           2,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 || files[0].FileID != "checksum-image-001" {
		t.Fatalf("expected checksum-image-001, got %#v", files)
	}
}

func TestMemoryRepositoryFindByTagCountsDoesNotAccumulateCaseDuplicates(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID: "checksum-image-001",
			TagCounts: map[string]int{
				"Alectura_lathami": 1,
				"alectura_lathami": 1,
			},
		},
	})

	files, err := repo.FindByTagCounts(
		context.Background(),
		map[string]int{"alectura_lathami": 2},
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

func TestDynamoDBRepositoryFindByTagCountsNormalisesHistoricalKeys(t *testing.T) {
	uppercaseItem := mustMarshalMediaFile(t, model.MediaFile{
		FileID: "checksum-image-001",
		TagCounts: map[string]int{
			"Alectura_lathami": 1,
			"magpie":           2,
		},
	})
	lowercaseItem := mustMarshalMediaFile(t, model.MediaFile{
		FileID: "checksum-image-002",
		TagCounts: map[string]int{
			"alectura_lathami": 1,
		},
	})

	newRepo := func() *DynamoDBRepository {
		return NewDynamoDBRepository(
			&fakeDynamoDBClient{
				pages: []*dynamodb.ScanOutput{
					{
						Items: []map[string]types.AttributeValue{
							uppercaseItem,
							lowercaseItem,
						},
					},
				},
			},
			"test-table",
		)
	}

	files, err := newRepo().FindByTagCounts(
		context.Background(),
		map[string]int{"alectura_lathami": 1},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	files, err = newRepo().FindByTagCounts(
		context.Background(),
		map[string]int{"Alectura_lathami": 1},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	files, err = newRepo().FindByTagCounts(
		context.Background(),
		map[string]int{
			"Alectura_lathami": 1,
			"magpie":           2,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 || files[0].FileID != "checksum-image-001" {
		t.Fatalf("expected checksum-image-001, got %#v", files)
	}
}

func TestDynamoDBRepositoryFindByTagCountsDoesNotMatchInsufficientMixedCaseCount(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID: "checksum-image-001",
		TagCounts: map[string]int{
			"Alectura_lathami": 1,
			"magpie":           1,
		},
	})

	repo := NewDynamoDBRepository(
		&fakeDynamoDBClient{
			pages: []*dynamodb.ScanOutput{
				{
					Items: []map[string]types.AttributeValue{
						item,
					},
				},
			},
		},
		"test-table",
	)

	files, err := repo.FindByTagCounts(
		context.Background(),
		map[string]int{
			"alectura_lathami": 1,
			"magpie":           2,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}
