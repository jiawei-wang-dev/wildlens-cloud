package repository

import (
	"strings"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

// newFileIDSet trims file IDs, removes empty values and removes duplicates.
func newFileIDSet(fileIDs []string) map[string]struct{} {
	results := make(map[string]struct{})

	for _, fileID := range fileIDs {
		fileID = strings.TrimSpace(fileID)

		if fileID == "" {
			continue
		}

		results[fileID] = struct{}{}
	}

	return results
}

// findMediaFilesByURLs returns metadata whose original or thumbnail URL matches.
func findMediaFilesByURLs(
	files []model.MediaFile,
	urls []string,
) []model.MediaFile {
	targetURLs := newURLSet(urls)
	results := make([]model.MediaFile, 0)

	if len(targetURLs) == 0 {
		return results
	}

	for _, file := range files {
		if matchesMediaURL(file, targetURLs) {
			results = append(results, file)
		}
	}

	return results
}
