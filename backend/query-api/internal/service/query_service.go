package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/storage"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/repository"
)

var (
	ErrSpeciesRequired         = errors.New("species is required")
	ErrTagsRequired            = errors.New("at least one tag is required")
	ErrInvalidMinimumCount     = errors.New("minimum tag count must be greater than zero")
	ErrThumbnailURLRequired    = errors.New("thumbnail_url is required")
	ErrURLsRequired            = errors.New("at least one URL is required")
	ErrTagOperationRequired    = errors.New("operation is required")
	ErrInvalidObservationLimit = errors.New(
		"limit must be between 1 and 50",
	)
)

// QueryService contains media query business logic.
type QueryService struct {
	repo      repository.MediaRepository
	urlSigner storage.MediaURLSigner
}

// NewQueryService creates a query service with a local URL signer.
func NewQueryService(
	repo repository.MediaRepository,
) *QueryService {
	return NewQueryServiceWithURLSigner(
		repo,
		storage.NewStaticURLSigner(""),
	)
}

// NewQueryServiceWithURLSigner creates a query service with a custom signer.
func NewQueryServiceWithURLSigner(
	repo repository.MediaRepository,
	urlSigner storage.MediaURLSigner,
) *QueryService {
	if urlSigner == nil {
		urlSigner = storage.NewStaticURLSigner("")
	}

	return &QueryService{
		repo:      repo,
		urlSigner: urlSigner,
	}
}

func (s *QueryService) FindBySpecies(
	ctx context.Context,
	species string,
) ([]model.MediaFile, error) {
	species = strings.TrimSpace(species)

	if species == "" {
		return nil, ErrSpeciesRequired
	}

	return s.repo.FindBySpecies(ctx, species)
}

func (s *QueryService) FindByTagCounts(
	ctx context.Context,
	required map[string]int,
) ([]model.MediaFile, error) {
	if len(required) == 0 {
		return nil, ErrTagsRequired
	}

	for _, count := range required {
		if count <= 0 {
			return nil, ErrInvalidMinimumCount
		}
	}

	return s.repo.FindByTagCounts(ctx, required)
}

func (s *QueryService) FindOriginalByThumbnailURL(
	ctx context.Context,
	thumbnailURL string,
) (string, error) {
	thumbnailURL = strings.TrimSpace(thumbnailURL)

	if thumbnailURL == "" {
		return "", ErrThumbnailURLRequired
	}

	return s.repo.FindOriginalByThumbnailURL(ctx, thumbnailURL)
}

// UpdateTags validates and applies a bulk tag update.
func (s *QueryService) UpdateTags(
	ctx context.Context,
	urls []string,
	tags []string,
	operation *int,
) ([]model.MediaFile, error) {
	cleanURLs := make([]string, 0, len(urls))

	for _, url := range urls {
		url = strings.TrimSpace(url)

		if url != "" {
			cleanURLs = append(cleanURLs, url)
		}
	}

	if len(cleanURLs) == 0 {
		return nil, ErrURLsRequired
	}

	cleanTags := make([]string, 0, len(tags))

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)

		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}

	if len(cleanTags) == 0 {
		return nil, ErrTagsRequired
	}

	if operation == nil {
		return nil, ErrTagOperationRequired
	}

	if *operation != repository.TagOperationRemove &&
		*operation != repository.TagOperationAdd {
		return nil, repository.ErrInvalidTagOperation
	}

	return s.repo.UpdateTags(
		ctx,
		cleanURLs,
		cleanTags,
		*operation,
	)
}

// DeleteFiles validates URLs and removes matched media records.
func (s *QueryService) DeleteFiles(
	ctx context.Context,
	urls []string,
) ([]string, error) {
	cleanURLs := make([]string, 0, len(urls))
	seenURLs := make(map[string]struct{})

	for _, url := range urls {
		url = strings.TrimSpace(url)

		if url == "" {
			continue
		}

		if _, exists := seenURLs[url]; exists {
			continue
		}

		seenURLs[url] = struct{}{}
		cleanURLs = append(cleanURLs, url)
	}

	if len(cleanURLs) == 0 {
		return nil, ErrURLsRequired
	}

	files, err := s.repo.FindByURLs(ctx, cleanURLs)
	if err != nil {
		return nil, err
	}

	fileIDs := make([]string, 0, len(files))

	for _, file := range files {
		fileID := strings.TrimSpace(file.FileID)

		if fileID != "" {
			fileIDs = append(fileIDs, fileID)
		}
	}

	deletedFiles, err := s.repo.DeleteFiles(ctx, fileIDs)
	if err != nil {
		return nil, err
	}

	deletedFileIDs := make([]string, 0, len(deletedFiles))

	for _, file := range deletedFiles {
		deletedFileIDs = append(deletedFileIDs, file.FileID)
	}

	return deletedFileIDs, nil
}

// ListObservations validates filters and adds temporary display URLs.
func (s *QueryService) ListObservations(
	ctx context.Context,
	limit int,
	nextToken string,
	species string,
	tags []string,
	fileType string,
	status string,
) (repository.ObservationPage, error) {
	if limit < 1 || limit > repository.MaxObservationLimit {
		return repository.ObservationPage{},
			ErrInvalidObservationLimit
	}

	page, err := s.repo.ListObservations(
		ctx,
		repository.ObservationListOptions{
			Limit:     limit,
			NextToken: strings.TrimSpace(nextToken),
			Species:   strings.TrimSpace(species),
			Tags:      cleanObservationTags(tags),
			FileType:  strings.TrimSpace(fileType),
			Status:    strings.TrimSpace(status),
		},
	)
	if err != nil {
		return repository.ObservationPage{}, err
	}

	for index := range page.Items {
		file := &page.Items[index]

		if strings.TrimSpace(file.ObjectPath) != "" {
			file.FileDownloadURL, err =
				s.urlSigner.SignGetObjectURL(
					ctx,
					file.Bucket,
					file.ObjectPath,
				)
			if err != nil {
				return repository.ObservationPage{},
					fmt.Errorf(
						"sign file download URL: %w",
						err,
					)
			}
		}

		if strings.TrimSpace(file.ThumbnailObjectPath) != "" {
			file.ThumbnailDisplayURL, err =
				s.urlSigner.SignGetObjectURL(
					ctx,
					file.Bucket,
					file.ThumbnailObjectPath,
				)
			if err != nil {
				return repository.ObservationPage{},
					fmt.Errorf(
						"sign thumbnail display URL: %w",
						err,
					)
			}
		}
	}

	return page, nil
}

func cleanObservationTags(tags []string) []string {
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
