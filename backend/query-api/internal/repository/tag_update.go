package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

const (
	// TagOperationRemove removes tags from matched media files.
	TagOperationRemove = 0

	// TagOperationAdd adds tags to matched media files.
	TagOperationAdd = 1
)

var ErrInvalidTagOperation = errors.New(
	"operation must be 0 (remove) or 1 (add)",
)

// applyTagUpdate modifies tags and counts for one media file.
func applyTagUpdate(
	file *model.MediaFile,
	tags []string,
	operation int,
) error {
	normalisedTags := normaliseTags(tags)

	switch operation {
	case TagOperationAdd:
		addTags(file, normalisedTags)

	case TagOperationRemove:
		removeTags(file, normalisedTags)

	default:
		return ErrInvalidTagOperation
	}

	file.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	return nil
}

func addTags(file *model.MediaFile, tags []string) {
	if file.TagCounts == nil {
		file.TagCounts = make(map[string]int)
	}

	existingTags := make(map[string]struct{})
	normalisedExistingTags := make([]string, 0, len(file.Tags))

	for _, tag := range file.Tags {
		tag = strings.ToLower(strings.TrimSpace(tag))

		if tag == "" {
			continue
		}

		if _, exists := existingTags[tag]; exists {
			continue
		}

		existingTags[tag] = struct{}{}
		normalisedExistingTags = append(normalisedExistingTags, tag)
	}

	for _, tag := range tags {
		if _, exists := existingTags[tag]; !exists {
			existingTags[tag] = struct{}{}
			normalisedExistingTags = append(normalisedExistingTags, tag)
		}

		if _, exists := file.TagCounts[tag]; !exists {
			file.TagCounts[tag] = 1
		}
	}

	file.Tags = normalisedExistingTags
}

func removeTags(file *model.MediaFile, tags []string) {
	tagsToRemove := make(map[string]struct{})

	for _, tag := range tags {
		tagsToRemove[tag] = struct{}{}
	}

	remainingTags := make([]string, 0, len(file.Tags))

	for _, tag := range file.Tags {
		tag = strings.ToLower(strings.TrimSpace(tag))

		if tag == "" {
			continue
		}

		if _, remove := tagsToRemove[tag]; remove {
			continue
		}

		remainingTags = append(remainingTags, tag)
	}

	file.Tags = remainingTags

	for tag := range tagsToRemove {
		delete(file.TagCounts, tag)
	}
}

func normaliseTags(tags []string) []string {
	results := make([]string, 0, len(tags))
	seen := make(map[string]struct{})

	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))

		if tag == "" {
			continue
		}

		if _, exists := seen[tag]; exists {
			continue
		}

		seen[tag] = struct{}{}
		results = append(results, tag)
	}

	return results
}

func newURLSet(urls []string) map[string]struct{} {
	results := make(map[string]struct{})

	for _, url := range urls {
		url = strings.TrimSpace(url)

		if url == "" {
			continue
		}

		results[url] = struct{}{}
	}

	return results
}

func matchesMediaURL(
	file model.MediaFile,
	urls map[string]struct{},
) bool {
	if _, exists := urls[file.FileURL]; exists {
		return true
	}

	if file.ThumbnailURL != "" {
		if _, exists := urls[file.ThumbnailURL]; exists {
			return true
		}
	}

	return false
}
