package repository

import (
	"net/url"
	"strings"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

func matchesThumbnailLookup(
	file model.MediaFile,
	thumbnailURL string,
) bool {
	thumbnailURL = strings.TrimSpace(thumbnailURL)

	if thumbnailURL == "" {
		return false
	}

	if file.ThumbnailURL == thumbnailURL {
		return true
	}

	targetPaths := objectPathCandidates(thumbnailURL, file.Bucket)
	filePaths := make(map[string]struct{})

	for _, path := range objectPathCandidates(file.ThumbnailURL, file.Bucket) {
		filePaths[path] = struct{}{}
	}

	if strings.TrimSpace(file.ThumbnailObjectPath) != "" {
		filePaths[strings.TrimLeft(file.ThumbnailObjectPath, "/")] = struct{}{}
	}

	for _, path := range targetPaths {
		if _, exists := filePaths[path]; exists {
			return true
		}
	}

	return false
}

func objectPathCandidates(
	value string,
	bucket string,
) []string {
	value = strings.TrimSpace(value)
	bucket = strings.TrimSpace(bucket)

	if value == "" {
		return []string{}
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return []string{strings.TrimLeft(strings.Split(value, "?")[0], "/")}
	}

	candidates := make([]string, 0, 2)
	path := strings.TrimLeft(parsed.Path, "/")

	if parsed.Scheme == "s3" {
		path = strings.TrimLeft(parsed.Path, "/")
	}

	if path != "" {
		candidates = append(candidates, path)
	}

	if bucket != "" && strings.HasPrefix(path, bucket+"/") {
		candidates = append(candidates, strings.TrimPrefix(path, bucket+"/"))
	}

	return deduplicateStrings(candidates)
}

func deduplicateStrings(values []string) []string {
	results := make([]string, 0, len(values))
	seen := make(map[string]struct{})

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		results = append(results, value)
	}

	return results
}
