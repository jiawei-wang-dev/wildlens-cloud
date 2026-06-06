package repository

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

const (
	// DefaultObservationLimit is used when the client does not provide a limit.
	DefaultObservationLimit = 10

	// MaxObservationLimit prevents very large list responses.
	MaxObservationLimit = 50
)

var ErrInvalidNextToken = errors.New("invalid next_token")

// ObservationListOptions contains list filters and pagination settings.
type ObservationListOptions struct {
	Limit     int
	NextToken string
	Species   string
	Tags      []string
	FileType  string
	Status    string
}

// ObservationPage contains one paginated list response.
type ObservationPage struct {
	Items     []model.MediaFile
	NextToken string
	HasMore   bool
}

// paginateObservations filters, sorts and paginates media records.
func paginateObservations(
	files []model.MediaFile,
	options ObservationListOptions,
) (ObservationPage, error) {
	filteredFiles := filterObservations(files, options)

	sort.Slice(filteredFiles, func(left int, right int) bool {
		return compareObservationCreatedAt(
			filteredFiles[left],
			filteredFiles[right],
		)
	})

	offset, err := decodeNextToken(options.NextToken)
	if err != nil {
		return ObservationPage{}, err
	}

	if offset > len(filteredFiles) {
		return ObservationPage{}, ErrInvalidNextToken
	}

	end := offset + options.Limit

	if end > len(filteredFiles) {
		end = len(filteredFiles)
	}

	items := filteredFiles[offset:end]

	normaliseObservationCreatedAt(items)

	hasMore := end < len(filteredFiles)

	nextToken := ""

	if hasMore {
		nextToken = encodeNextToken(end)
	}

	return ObservationPage{
		Items:     items,
		NextToken: nextToken,
		HasMore:   hasMore,
	}, nil
}

func filterObservations(
	files []model.MediaFile,
	options ObservationListOptions,
) []model.MediaFile {
	species := strings.TrimSpace(options.Species)
	tags := normaliseTags(options.Tags)
	fileType := strings.ToLower(strings.TrimSpace(options.FileType))
	status := strings.ToLower(strings.TrimSpace(options.Status))

	results := make([]model.MediaFile, 0)

	for _, file := range files {
		if species != "" &&
			strings.TrimSpace(file.PrimarySpecies) != species {
			continue
		}

		if !matchesRequiredObservationTags(file, tags) {
			continue
		}

		if fileType != "" &&
			strings.ToLower(strings.TrimSpace(file.FileType)) != fileType {
			continue
		}

		if status != "" &&
			strings.ToLower(strings.TrimSpace(file.Status)) != status {
			continue
		}

		results = append(results, file)
	}

	return results
}

func matchesRequiredObservationTags(
	file model.MediaFile,
	requiredTags []string,
) bool {
	for _, tag := range requiredTags {
		if hasObservationTag(file, tag) {
			continue
		}

		return false
	}

	return true
}

func hasObservationTag(
	file model.MediaFile,
	tag string,
) bool {
	if file.TagCounts[tag] >= 1 {
		return true
	}

	for _, existingTag := range file.Tags {
		existingTag = strings.ToLower(strings.TrimSpace(existingTag))

		if existingTag == tag {
			return true
		}
	}

	return false
}

func compareObservationCreatedAt(
	left model.MediaFile,
	right model.MediaFile,
) bool {
	leftTime, leftOK := parseObservationCreatedAt(left.CreatedAt)
	rightTime, rightOK := parseObservationCreatedAt(right.CreatedAt)

	if leftOK && rightOK {
		if !leftTime.Equal(rightTime) {
			return leftTime.After(rightTime)
		}

		return left.FileID < right.FileID
	}

	if leftOK != rightOK {
		return leftOK
	}

	return left.FileID < right.FileID
}

func normaliseObservationCreatedAt(files []model.MediaFile) {
	for index := range files {
		createdAt, ok := parseObservationCreatedAt(files[index].CreatedAt)

		if ok {
			files[index].CreatedAt = createdAt.UTC().Format(time.RFC3339)
		}
	}
}

func parseObservationCreatedAt(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)

	if value == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)

		if err == nil {
			return parsed.UTC(), true
		}
	}

	return time.Time{}, false
}

func encodeNextToken(offset int) string {
	value := strconv.Itoa(offset)

	return base64.RawURLEncoding.EncodeToString(
		[]byte(value),
	)
}

func decodeNextToken(token string) (int, error) {
	token = strings.TrimSpace(token)

	if token == "" {
		return 0, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, ErrInvalidNextToken
	}

	offset, err := strconv.Atoi(string(decoded))
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("%w: invalid offset", ErrInvalidNextToken)
	}

	return offset, nil
}
