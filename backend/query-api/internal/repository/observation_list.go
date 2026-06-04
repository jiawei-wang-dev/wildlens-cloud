package repository

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

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
		return filteredFiles[left].FileID < filteredFiles[right].FileID
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
	species := strings.ToLower(strings.TrimSpace(options.Species))
	fileType := strings.ToLower(strings.TrimSpace(options.FileType))
	status := strings.ToLower(strings.TrimSpace(options.Status))

	results := make([]model.MediaFile, 0)

	for _, file := range files {
		if species != "" && file.TagCounts[species] < 1 {
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
