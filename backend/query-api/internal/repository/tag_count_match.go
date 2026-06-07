package repository

import (
	"strings"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

func matchesTagCountMinimums(
	file model.MediaFile,
	required map[string]int,
) bool {
	normalisedCounts := normaliseTagCountKeys(file.TagCounts)

	for tag, minimumCount := range required {
		tag = strings.ToLower(strings.TrimSpace(tag))

		if tag == "" {
			continue
		}

		if normalisedCounts[tag] < minimumCount {
			return false
		}
	}

	return true
}

func normaliseTagCountKeys(
	tagCounts map[string]int,
) map[string]int {
	results := make(map[string]int, len(tagCounts))

	for tag, count := range tagCounts {
		tag = strings.ToLower(strings.TrimSpace(tag))

		if tag == "" {
			continue
		}

		if existingCount, exists := results[tag]; exists &&
			existingCount >= count {
			continue
		}

		results[tag] = count
	}

	return results
}
