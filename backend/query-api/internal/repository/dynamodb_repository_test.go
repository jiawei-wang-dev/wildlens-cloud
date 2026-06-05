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
	pages        []*dynamodb.ScanOutput
	err          error
	calls        int
	updateInputs []*dynamodb.UpdateItemInput
	updateErr    error
	deleteInputs []*dynamodb.DeleteItemInput
	deleteErr    error
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

func (f *fakeDynamoDBClient) UpdateItem(
	_ context.Context,
	params *dynamodb.UpdateItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.UpdateItemOutput, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}

	f.updateInputs = append(f.updateInputs, params)

	return &dynamodb.UpdateItemOutput{}, nil
}

func (f *fakeDynamoDBClient) DeleteItem(
	_ context.Context,
	params *dynamodb.DeleteItemInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.DeleteItemOutput, error) {
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}

	f.deleteInputs = append(f.deleteInputs, params)

	return &dynamodb.DeleteItemOutput{}, nil
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

func TestDynamoDBRepositoryFindByURLsMatchesOriginalURL(t *testing.T) {
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

	files, err := repo.FindByURLs(
		context.Background(),
		[]string{"s3://bucket/originals/koala.jpg"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].FileID != "checksum-image-001" {
		t.Fatalf("unexpected file ID: %s", files[0].FileID)
	}
}

func TestDynamoDBRepositoryFindByURLsMatchesThumbnailURL(t *testing.T) {
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

	files, err := repo.FindByURLs(
		context.Background(),
		[]string{"s3://bucket/thumbnails/koala.jpg"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].FileID != "checksum-image-001" {
		t.Fatalf("unexpected file ID: %s", files[0].FileID)
	}
}

func TestDynamoDBRepositoryFindByURLsReturnsScanError(t *testing.T) {
	client := &fakeDynamoDBClient{
		err: errors.New("scan failed"),
	}

	repo := NewDynamoDBRepository(client, "test-table")

	_, err := repo.FindByURLs(
		context.Background(),
		[]string{"s3://bucket/originals/koala.jpg"},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
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

func TestDynamoDBRepositoryUpdateTagsPersistsChanges(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID:       "checksum-image-001",
		FileURL:      "s3://bucket/originals/koala.jpg",
		ThumbnailURL: "s3://bucket/thumbnails/koala.jpg",
		Tags:         []string{"koala"},
		TagCounts: map[string]int{
			"koala": 3,
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

	files, err := repo.UpdateTags(
		context.Background(),
		[]string{"s3://bucket/thumbnails/koala.jpg"},
		[]string{"reviewed"},
		TagOperationAdd,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 updated file, got %d", len(files))
	}

	if files[0].TagCounts["reviewed"] != 1 {
		t.Fatalf(
			"expected reviewed count 1, got %d",
			files[0].TagCounts["reviewed"],
		)
	}

	if len(client.updateInputs) != 1 {
		t.Fatalf(
			"expected 1 UpdateItem call, got %d",
			len(client.updateInputs),
		)
	}

	input := client.updateInputs[0]

	key, exists := input.Key["file_id"]
	if !exists {
		t.Fatal("expected file_id key")
	}

	fileID, ok := key.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected string file_id key")
	}

	if fileID.Value != "checksum-image-001" {
		t.Fatalf("unexpected file_id: %s", fileID.Value)
	}
}

func TestDynamoDBRepositoryUpdateTagsReturnsUpdateError(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID:  "checksum-image-001",
		FileURL: "s3://bucket/originals/koala.jpg",
		Tags:    []string{"koala"},
		TagCounts: map[string]int{
			"koala": 1,
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
		updateErr: errors.New("update failed"),
	}

	repo := NewDynamoDBRepository(client, "test-table")

	_, err := repo.UpdateTags(
		context.Background(),
		[]string{"s3://bucket/originals/koala.jpg"},
		[]string{"reviewed"},
		TagOperationAdd,
	)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestDynamoDBRepositoryDeleteFilesPersistsDeletion(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID:              "checksum-image-001",
		FileURL:             "s3://bucket/originals/koala.jpg",
		ObjectPath:          "media/originals/koala.jpg",
		ThumbnailObjectPath: "media/thumbnails/koala.jpg",
		Status:              "ready",
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

	files, err := repo.DeleteFiles(
		context.Background(),
		[]string{"checksum-image-001"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 deleted file, got %d", len(files))
	}

	if files[0].FileID != "checksum-image-001" {
		t.Fatalf("unexpected deleted file ID: %s", files[0].FileID)
	}

	if len(client.deleteInputs) != 1 {
		t.Fatalf(
			"expected 1 DeleteItem call, got %d",
			len(client.deleteInputs),
		)
	}

	input := client.deleteInputs[0]

	key, exists := input.Key["file_id"]
	if !exists {
		t.Fatal("expected file_id key")
	}

	fileID, ok := key.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected string file_id key")
	}

	if fileID.Value != "checksum-image-001" {
		t.Fatalf("unexpected file_id: %s", fileID.Value)
	}
}

func TestDynamoDBRepositoryDeleteFilesIgnoresUnknownID(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID:  "checksum-image-001",
		FileURL: "s3://bucket/originals/koala.jpg",
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

	files, err := repo.DeleteFiles(
		context.Background(),
		[]string{"missing-file-id"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 deleted files, got %d", len(files))
	}

	if len(client.deleteInputs) != 0 {
		t.Fatalf(
			"expected 0 DeleteItem calls, got %d",
			len(client.deleteInputs),
		)
	}
}

func TestDynamoDBRepositoryDeleteFilesReturnsDeleteError(t *testing.T) {
	item := mustMarshalMediaFile(t, model.MediaFile{
		FileID:  "checksum-image-001",
		FileURL: "s3://bucket/originals/koala.jpg",
	})

	client := &fakeDynamoDBClient{
		pages: []*dynamodb.ScanOutput{
			{
				Items: []map[string]types.AttributeValue{
					item,
				},
			},
		},
		deleteErr: errors.New("delete failed"),
	}

	repo := NewDynamoDBRepository(client, "test-table")

	_, err := repo.DeleteFiles(
		context.Background(),
		[]string{"checksum-image-001"},
	)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}
