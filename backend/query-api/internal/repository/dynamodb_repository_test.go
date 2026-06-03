package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

// fakeDynamoDBClient simulates DynamoDB Scan responses without AWS access.
type fakeDynamoDBClient struct {
	pages []*dynamodb.ScanOutput
	err   error
	calls int
}

func (f *fakeDynamoDBClient) Scan(
	_ context.Context,
	_ *dynamodb.ScanInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.ScanOutput, error) {
	if f.err != nil {
		return nil, f.err
	}

	if f.calls >= len(f.pages) {
		return &dynamodb.ScanOutput{}, nil
	}

	page := f.pages[f.calls]
	f.calls++

	return page, nil
}

func TestDynamoDBRepositoryFindBySpeciesReadsAllPages(t *testing.T) {
	imageItem := mustMarshalMediaFile(t, model.MediaFile{
		FileID:   "checksum-image-001",
		FileType: "image",
		FileURL:  "s3://bucket/originals/koala.jpg",
		Tags:     []string{"koala"},
		TagCounts: map[string]int{
			"koala": 1,
		},
		Status: "ready",
	})

	videoItem := mustMarshalMediaFile(t, model.MediaFile{
		FileID:   "checksum-video-001",
		FileType: "video",
		FileURL:  "s3://bucket/originals/wombat.mp4",
		Tags:     []string{"wombat"},
		TagCounts: map[string]int{
			"wombat": 2,
		},
		Status: "ready",
	})

	client := &fakeDynamoDBClient{
		pages: []*dynamodb.ScanOutput{
			{
				Items: []map[string]types.AttributeValue{
					imageItem,
				},
				LastEvaluatedKey: map[string]types.AttributeValue{
					"file_id": &types.AttributeValueMemberS{
						Value: "checksum-image-001",
					},
				},
			},
			{
				Items: []map[string]types.AttributeValue{
					videoItem,
				},
			},
		},
	}

	repo := NewDynamoDBRepository(client, "test-table")

	files, err := repo.FindBySpecies(
		context.Background(),
		"wombat",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].FileID != "checksum-video-001" {
		t.Fatalf("unexpected file ID: %s", files[0].FileID)
	}

	if client.calls != 2 {
		t.Fatalf("expected 2 Scan calls, got %d", client.calls)
	}
}

func TestDynamoDBRepositoryFindByTagCountsUsesAND(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID:  "checksum-image-001",
		FileURL: "s3://bucket/originals/koala.jpg",
		TagCounts: map[string]int{
			"koala":  3,
			"magpie": 1,
		},
		Status: "ready",
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
			"wombat": 1,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestDynamoDBRepositoryFindOriginalByThumbnailURL(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID:       "checksum-image-001",
		FileURL:      "s3://bucket/originals/koala.jpg",
		ThumbnailURL: "s3://bucket/thumbnails/koala.jpg",
		Status:       "ready",
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

	fileURL, err := repo.FindOriginalByThumbnailURL(
		context.Background(),
		"s3://bucket/thumbnails/koala.jpg",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "s3://bucket/originals/koala.jpg"

	if fileURL != expected {
		t.Fatalf("expected %q, got %q", expected, fileURL)
	}
}

func TestDynamoDBRepositoryReturnsScanError(t *testing.T) {
	client := &fakeDynamoDBClient{
		err: errors.New("scan failed"),
	}

	repo := NewDynamoDBRepository(client, "test-table")

	_, err := repo.FindBySpecies(
		context.Background(),
		"koala",
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func mustMarshalMediaFile(
	t *testing.T,
	file model.MediaFile,
) map[string]types.AttributeValue {
	t.Helper()

	item, err := attributevalue.MarshalMap(file)
	if err != nil {
		t.Fatalf("failed to marshal media file: %v", err)
	}

	return item
}
