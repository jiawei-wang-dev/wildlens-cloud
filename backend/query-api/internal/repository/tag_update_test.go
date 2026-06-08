package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

func TestMemoryRepositoryUpdateTagsAddsNewTag(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:       "checksum-image-001",
			FileURL:      "s3://bucket/originals/koala.jpg",
			ThumbnailURL: "s3://bucket/thumbnails/koala.jpg",
			Tags:         []string{"koala"},
			TagCounts: map[string]int{
				"koala": 3,
			},
			Status: "ready",
		},
	})

	updatedFiles, err := repo.UpdateTags(
		context.Background(),
		[]string{"s3://bucket/thumbnails/koala.jpg"},
		[]string{" reviewed ", "koala", "reviewed"},
		TagOperationAdd,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(updatedFiles) != 1 {
		t.Fatalf("expected 1 updated file, got %d", len(updatedFiles))
	}

	file := updatedFiles[0]

	if len(file.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(file.Tags))
	}

	if file.TagCounts["koala"] != 3 {
		t.Fatalf(
			"expected existing koala count to remain 3, got %d",
			file.TagCounts["koala"],
		)
	}

	if file.TagCounts["reviewed"] != 1 {
		t.Fatalf(
			"expected reviewed count 1, got %d",
			file.TagCounts["reviewed"],
		)
	}

	if file.UpdatedAt == "" {
		t.Fatal("expected updated_at to be set")
	}
}

func TestMemoryRepositoryUpdateTagsRemovesExistingTag(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:  "checksum-image-001",
			FileURL: "s3://bucket/originals/koala.jpg",
			Tags:    []string{"koala", "reviewed"},
			TagCounts: map[string]int{
				"koala":    3,
				"reviewed": 1,
			},
			Status: "ready",
		},
	})

	updatedFiles, err := repo.UpdateTags(
		context.Background(),
		[]string{"s3://bucket/originals/koala.jpg"},
		[]string{"reviewed", "missing"},
		TagOperationRemove,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(updatedFiles) != 1 {
		t.Fatalf("expected 1 updated file, got %d", len(updatedFiles))
	}

	file := updatedFiles[0]

	if len(file.Tags) != 1 || file.Tags[0] != "koala" {
		t.Fatalf("unexpected remaining tags: %v", file.Tags)
	}

	if _, exists := file.TagCounts["reviewed"]; exists {
		t.Fatal("expected reviewed count to be removed")
	}
}

func TestMemoryRepositoryUpdateTagsIgnoresUnknownURL(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:  "checksum-image-001",
			FileURL: "s3://bucket/originals/koala.jpg",
		},
	})

	updatedFiles, err := repo.UpdateTags(
		context.Background(),
		[]string{"s3://bucket/originals/missing.jpg"},
		[]string{"reviewed"},
		TagOperationAdd,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(updatedFiles) != 0 {
		t.Fatalf("expected 0 updated files, got %d", len(updatedFiles))
	}
}

func TestMemoryRepositoryUpdateTagsRejectsInvalidOperation(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:  "checksum-image-001",
			FileURL: "s3://bucket/originals/koala.jpg",
		},
	})

	_, err := repo.UpdateTags(
		context.Background(),
		[]string{"s3://bucket/originals/koala.jpg"},
		[]string{"reviewed"},
		99,
	)

	if !errors.Is(err, ErrInvalidTagOperation) {
		t.Fatalf(
			"expected ErrInvalidTagOperation, got %v",
			err,
		)
	}
}
