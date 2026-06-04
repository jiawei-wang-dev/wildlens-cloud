package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

var ErrMediaNotFound = errors.New("media file not found")

// MemoryRepository is a temporary in-memory database.
// Replace it with Firestore or DynamoDB later.
type MemoryRepository struct {
	files []model.MediaFile
}

func NewMemoryRepository(files []model.MediaFile) *MemoryRepository {
	return &MemoryRepository{files: files}
}

// FindBySpecies returns files containing at least one requested species.
func (r *MemoryRepository) FindBySpecies(
	_ context.Context,
	species string,
) ([]model.MediaFile, error) {
	species = strings.ToLower(strings.TrimSpace(species))
	results := make([]model.MediaFile, 0)

	for _, file := range r.files {
		if file.TagCounts[species] >= 1 {
			results = append(results, file)
		}
	}

	return results, nil
}

// FindByTagCounts applies logical AND between all minimum-count conditions.
func (r *MemoryRepository) FindByTagCounts(
	_ context.Context,
	required map[string]int,
) ([]model.MediaFile, error) {
	results := make([]model.MediaFile, 0)

	for _, file := range r.files {
		matched := true

		for tag, minimumCount := range required {
			tag = strings.ToLower(strings.TrimSpace(tag))

			if file.TagCounts[tag] < minimumCount {
				matched = false
				break
			}
		}

		if matched {
			results = append(results, file)
		}
	}

	return results, nil
}

// FindOriginalByThumbnailURL maps a thumbnail URL to the original file URL.
func (r *MemoryRepository) FindOriginalByThumbnailURL(
	_ context.Context,
	thumbnailURL string,
) (string, error) {
	thumbnailURL = strings.TrimSpace(thumbnailURL)

	for _, file := range r.files {
		if file.ThumbnailURL == thumbnailURL {
			return file.FileURL, nil
		}
	}

	return "", ErrMediaNotFound
}

// UpdateTags adds or removes tags for media files matching the supplied URLs.
// UpdateTags adds or removes tags for media files matching the supplied URLs.
func (r *MemoryRepository) UpdateTags(
	_ context.Context,
	urls []string,
	tags []string,
	operation int,
) ([]model.MediaFile, error) {
	if err := validateTagOperation(operation); err != nil {
		return nil, err
	}

	targetURLs := newURLSet(urls)
	updatedFiles := make([]model.MediaFile, 0)

	for index := range r.files {
		if !matchesMediaURL(r.files[index], targetURLs) {
			continue
		}

		if err := applyTagUpdate(
			&r.files[index],
			tags,
			operation,
		); err != nil {
			return nil, err
		}

		updatedFiles = append(updatedFiles, r.files[index])
	}

	return updatedFiles, nil
}

// DeleteFiles removes media records matching the supplied file IDs.
func (r *MemoryRepository) DeleteFiles(
	_ context.Context,
	fileIDs []string,
) ([]model.MediaFile, error) {
	targetIDs := newFileIDSet(fileIDs)

	deletedFiles := make([]model.MediaFile, 0)
	remainingFiles := make([]model.MediaFile, 0, len(r.files))

	for _, file := range r.files {
		if _, exists := targetIDs[file.FileID]; exists {
			deletedFiles = append(deletedFiles, file)
			continue
		}

		remainingFiles = append(remainingFiles, file)
	}

	r.files = remainingFiles

	return deletedFiles, nil
}
