package repository

import (
	"context"
	"testing"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

func TestMemoryRepositoryFindByURLsMatchesOriginalURL(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:       "checksum-image-001",
			FileURL:      "s3://bucket/originals/koala.jpg",
			ThumbnailURL: "s3://bucket/thumbnails/koala.jpg",
		},
		{
			FileID:  "checksum-video-001",
			FileURL: "s3://bucket/originals/wombat.mp4",
		},
	})

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

func TestMemoryRepositoryFindByURLsMatchesThumbnailURL(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:       "checksum-image-001",
			FileURL:      "s3://bucket/originals/koala.jpg",
			ThumbnailURL: "s3://bucket/thumbnails/koala.jpg",
		},
	})

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

func TestMemoryRepositoryFindByURLsIgnoresUnknownURL(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:  "checksum-image-001",
			FileURL: "s3://bucket/originals/koala.jpg",
		},
	})

	files, err := repo.FindByURLs(
		context.Background(),
		[]string{"s3://bucket/originals/missing.jpg"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestMemoryRepositoryFindByURLsDeduplicatesURLs(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:       "checksum-image-001",
			FileURL:      "s3://bucket/originals/koala.jpg",
			ThumbnailURL: "s3://bucket/thumbnails/koala.jpg",
		},
	})

	files, err := repo.FindByURLs(
		context.Background(),
		[]string{
			" s3://bucket/originals/koala.jpg ",
			"s3://bucket/originals/koala.jpg",
			"",
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestMemoryRepositoryDeleteFilesRemovesMatchedFile(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:  "checksum-image-001",
			FileURL: "s3://bucket/originals/koala.jpg",
		},
		{
			FileID:  "checksum-video-001",
			FileURL: "s3://bucket/originals/wombat.mp4",
		},
	})

	deletedFiles, err := repo.DeleteFiles(
		context.Background(),
		[]string{"checksum-image-001"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(deletedFiles) != 1 {
		t.Fatalf(
			"expected 1 deleted file, got %d",
			len(deletedFiles),
		)
	}

	if deletedFiles[0].FileID != "checksum-image-001" {
		t.Fatalf(
			"unexpected deleted file ID: %s",
			deletedFiles[0].FileID,
		)
	}

	if len(repo.files) != 1 {
		t.Fatalf(
			"expected 1 remaining file, got %d",
			len(repo.files),
		)
	}

	if repo.files[0].FileID != "checksum-video-001" {
		t.Fatalf(
			"unexpected remaining file ID: %s",
			repo.files[0].FileID,
		)
	}
}

func TestMemoryRepositoryDeleteFilesTrimsAndDeduplicatesIDs(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:  "checksum-image-001",
			FileURL: "s3://bucket/originals/koala.jpg",
		},
	})

	deletedFiles, err := repo.DeleteFiles(
		context.Background(),
		[]string{
			" checksum-image-001 ",
			"checksum-image-001",
			"",
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(deletedFiles) != 1 {
		t.Fatalf(
			"expected 1 deleted file, got %d",
			len(deletedFiles),
		)
	}

	if len(repo.files) != 0 {
		t.Fatalf(
			"expected 0 remaining files, got %d",
			len(repo.files),
		)
	}
}

func TestMemoryRepositoryDeleteFilesIgnoresUnknownIDs(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:  "checksum-image-001",
			FileURL: "s3://bucket/originals/koala.jpg",
		},
	})

	deletedFiles, err := repo.DeleteFiles(
		context.Background(),
		[]string{"missing-file-id"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(deletedFiles) != 0 {
		t.Fatalf(
			"expected 0 deleted files, got %d",
			len(deletedFiles),
		)
	}

	if len(repo.files) != 1 {
		t.Fatalf(
			"expected 1 remaining file, got %d",
			len(repo.files),
		)
	}
}

func TestMemoryRepositoryDeleteFilesHandlesEmptyIDs(t *testing.T) {
	repo := NewMemoryRepository([]model.MediaFile{
		{
			FileID:  "checksum-image-001",
			FileURL: "s3://bucket/originals/koala.jpg",
		},
	})

	deletedFiles, err := repo.DeleteFiles(
		context.Background(),
		[]string{},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(deletedFiles) != 0 {
		t.Fatalf(
			"expected 0 deleted files, got %d",
			len(deletedFiles),
		)
	}

	if len(repo.files) != 1 {
		t.Fatalf(
			"expected 1 remaining file, got %d",
			len(repo.files),
		)
	}
}
