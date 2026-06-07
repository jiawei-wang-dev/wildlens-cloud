package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/inference"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/repository"
)

type fakeMediaRepository struct {
	calls             *[]string
	findByURLsResult  []model.MediaFile
	findByTagsResult  []model.MediaFile
	findByTagsRequest map[string]int
	deleteFilesResult []model.MediaFile
	deleteFilesIDs    []string
	deleteFilesCalled bool
	updateTagsCalled  bool
}

func (f *fakeMediaRepository) FindBySpecies(
	_ context.Context,
	_ string,
) ([]model.MediaFile, error) {
	return nil, nil
}

func (f *fakeMediaRepository) FindByTagCounts(
	_ context.Context,
	required map[string]int,
) ([]model.MediaFile, error) {
	*f.calls = append(*f.calls, "FindByTagCounts")
	f.findByTagsRequest = make(map[string]int)

	for tag, count := range required {
		f.findByTagsRequest[tag] = count
	}

	return f.findByTagsResult, nil
}

func (f *fakeMediaRepository) FindOriginalByThumbnailURL(
	_ context.Context,
	_ string,
) (string, error) {
	return "", nil
}

func (f *fakeMediaRepository) FindByURLs(
	_ context.Context,
	_ []string,
) ([]model.MediaFile, error) {
	*f.calls = append(*f.calls, "FindByURLs")

	return f.findByURLsResult, nil
}

func (f *fakeMediaRepository) UpdateTags(
	_ context.Context,
	_ []string,
	_ []string,
	_ int,
) ([]model.MediaFile, error) {
	f.updateTagsCalled = true

	return nil, nil
}

func (f *fakeMediaRepository) DeleteFiles(
	_ context.Context,
	fileIDs []string,
) ([]model.MediaFile, error) {
	*f.calls = append(*f.calls, "DeleteFiles")
	f.deleteFilesCalled = true
	f.deleteFilesIDs = append([]string{}, fileIDs...)

	return f.deleteFilesResult, nil
}

func (f *fakeMediaRepository) ListObservations(
	_ context.Context,
	_ repository.ObservationListOptions,
) (repository.ObservationPage, error) {
	return repository.ObservationPage{}, nil
}

type fakeObjectDeleter struct {
	calls             *[]string
	files             []model.MediaFile
	err               error
	deleteMediaCalled bool
}

type fakeImageInferenceClient struct {
	result      inference.ImageResult
	err         error
	called      bool
	filename    string
	contentType string
	data        []byte
}

func (f *fakeImageInferenceClient) InferImage(
	_ context.Context,
	filename string,
	contentType string,
	data []byte,
) (inference.ImageResult, error) {
	f.called = true
	f.filename = filename
	f.contentType = contentType
	f.data = append([]byte{}, data...)

	return f.result, f.err
}

func (f *fakeObjectDeleter) DeleteMediaObjects(
	_ context.Context,
	files []model.MediaFile,
) error {
	*f.calls = append(*f.calls, "DeleteMediaObjects")
	f.deleteMediaCalled = true
	f.files = append([]model.MediaFile{}, files...)

	return f.err
}

func TestDeleteFilesDeletesObjectsBeforeMetadata(t *testing.T) {
	calls := make([]string, 0)
	matchedFiles := []model.MediaFile{
		{
			FileID:              "checksum-image-001",
			Bucket:              "wildlens-media",
			ObjectPath:          "media/originals/koala.jpg",
			ThumbnailObjectPath: "media/thumbnails/koala.jpg",
		},
	}
	repo := &fakeMediaRepository{
		calls:             &calls,
		findByURLsResult:  matchedFiles,
		deleteFilesResult: matchedFiles,
	}
	objectDeleter := &fakeObjectDeleter{
		calls: &calls,
	}
	queryService := NewQueryServiceWithDependencies(
		repo,
		nil,
		objectDeleter,
	)

	deletedFileIDs, err := queryService.DeleteFiles(
		context.Background(),
		[]string{"s3://wildlens-media/media/originals/koala.jpg"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertCallOrder(
		t,
		calls,
		[]string{
			"FindByURLs",
			"DeleteMediaObjects",
			"DeleteFiles",
		},
	)

	if len(repo.deleteFilesIDs) != 1 ||
		repo.deleteFilesIDs[0] != "checksum-image-001" {
		t.Fatalf("unexpected deleted file IDs: %v", repo.deleteFilesIDs)
	}

	if len(deletedFileIDs) != 1 ||
		deletedFileIDs[0] != "checksum-image-001" {
		t.Fatalf("unexpected response file IDs: %v", deletedFileIDs)
	}
}

func TestDeleteFilesDoesNotDeleteMetadataWhenObjectDeleteFails(t *testing.T) {
	calls := make([]string, 0)
	repo := &fakeMediaRepository{
		calls: &calls,
		findByURLsResult: []model.MediaFile{
			{
				FileID:     "checksum-image-001",
				Bucket:     "wildlens-media",
				ObjectPath: "media/originals/koala.jpg",
			},
		},
	}
	objectDeleter := &fakeObjectDeleter{
		calls: &calls,
		err:   errors.New("delete objects failed"),
	}
	queryService := NewQueryServiceWithDependencies(
		repo,
		nil,
		objectDeleter,
	)

	_, err := queryService.DeleteFiles(
		context.Background(),
		[]string{"s3://wildlens-media/media/originals/koala.jpg"},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	assertCallOrder(
		t,
		calls,
		[]string{
			"FindByURLs",
			"DeleteMediaObjects",
		},
	)

	if repo.deleteFilesCalled {
		t.Fatal("expected metadata deletion not to be called")
	}
}

func TestDeleteFilesIgnoresUnknownURLWithoutObjectDelete(t *testing.T) {
	calls := make([]string, 0)
	repo := &fakeMediaRepository{
		calls:            &calls,
		findByURLsResult: []model.MediaFile{},
	}
	objectDeleter := &fakeObjectDeleter{
		calls: &calls,
	}
	queryService := NewQueryServiceWithDependencies(
		repo,
		nil,
		objectDeleter,
	)

	deletedFileIDs, err := queryService.DeleteFiles(
		context.Background(),
		[]string{"s3://wildlens-media/media/originals/missing.jpg"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertCallOrder(t, calls, []string{"FindByURLs"})

	if len(deletedFileIDs) != 0 {
		t.Fatalf("expected no deleted file IDs, got %v", deletedFileIDs)
	}

	if objectDeleter.deleteMediaCalled {
		t.Fatal("expected object deleter not to be called")
	}

	if repo.deleteFilesCalled {
		t.Fatal("expected metadata deletion not to be called")
	}
}

func TestDeleteFilesPassesMatchedMetadataToObjectDeleter(t *testing.T) {
	calls := make([]string, 0)
	matchedFiles := []model.MediaFile{
		{
			FileID:              "checksum-image-001",
			Bucket:              "wildlens-media",
			ObjectPath:          "media/originals/koala.jpg",
			ThumbnailObjectPath: "media/thumbnails/koala.jpg",
		},
		{
			FileID:     "checksum-video-001",
			Bucket:     "wildlens-media",
			ObjectPath: "media/originals/wombat.mp4",
		},
	}
	repo := &fakeMediaRepository{
		calls:             &calls,
		findByURLsResult:  matchedFiles,
		deleteFilesResult: matchedFiles,
	}
	objectDeleter := &fakeObjectDeleter{
		calls: &calls,
	}
	queryService := NewQueryServiceWithDependencies(
		repo,
		nil,
		objectDeleter,
	)

	_, err := queryService.DeleteFiles(
		context.Background(),
		[]string{
			"s3://wildlens-media/media/originals/koala.jpg",
			"s3://wildlens-media/media/originals/wombat.mp4",
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(objectDeleter.files) != len(matchedFiles) {
		t.Fatalf(
			"expected %d files, got %d",
			len(matchedFiles),
			len(objectDeleter.files),
		)
	}

	for index, file := range matchedFiles {
		if objectDeleter.files[index].FileID != file.FileID {
			t.Fatalf(
				"expected object deleter file %d to be %s, got %s",
				index,
				file.FileID,
				objectDeleter.files[index].FileID,
			)
		}
	}
}

func TestQueryByFileUsesInferenceTagsToFindMatches(t *testing.T) {
	calls := make([]string, 0)
	repo := &fakeMediaRepository{
		calls: &calls,
		findByTagsResult: []model.MediaFile{
			{
				FileID:              "checksum-image-001",
				Bucket:              "wildlens-media",
				ObjectPath:          "media/originals/koala.jpg",
				ThumbnailObjectPath: "media/thumbnails/koala.jpg",
				TagCounts: map[string]int{
					"koala":  3,
					"magpie": 1,
				},
			},
		},
	}
	imageInference := &fakeImageInferenceClient{
		result: inference.ImageResult{
			Tags: []string{"koala", "magpie"},
		},
	}
	objectDeleter := &fakeObjectDeleter{
		calls: &calls,
	}
	queryService := NewQueryServiceWithAllDependencies(
		repo,
		nil,
		objectDeleter,
		imageInference,
	)

	response, err := queryService.QueryByFile(
		context.Background(),
		"query.jpg",
		"image/jpeg",
		[]byte("image bytes"),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertCallOrder(t, calls, []string{"FindByTagCounts"})

	if !imageInference.called {
		t.Fatal("expected inference client to be called")
	}

	if repo.findByTagsRequest["koala"] != 1 ||
		repo.findByTagsRequest["magpie"] != 1 {
		t.Fatalf("unexpected tag count query: %v", repo.findByTagsRequest)
	}

	if len(response.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(response.Items))
	}

	if response.Items[0].FileDownloadURL == "" {
		t.Fatal("expected file_download_url")
	}

	if response.Items[0].ThumbnailDisplayURL == "" {
		t.Fatal("expected thumbnail_display_url")
	}

	if repo.updateTagsCalled || repo.deleteFilesCalled {
		t.Fatal("expected no repository writes")
	}

	if objectDeleter.deleteMediaCalled {
		t.Fatal("expected no S3 object deletion")
	}
}

func TestQueryByFileUsesInferenceTagCountsWithAND(t *testing.T) {
	calls := make([]string, 0)
	repo := &fakeMediaRepository{
		calls: &calls,
	}
	imageInference := &fakeImageInferenceClient{
		result: inference.ImageResult{
			TagCounts: map[string]int{
				"koala":  3,
				"magpie": 1,
			},
		},
	}
	queryService := NewQueryServiceWithAllDependencies(
		repo,
		nil,
		nil,
		imageInference,
	)

	_, err := queryService.QueryByFile(
		context.Background(),
		"query.png",
		"image/png",
		[]byte("image bytes"),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.findByTagsRequest["koala"] != 3 ||
		repo.findByTagsRequest["magpie"] != 1 {
		t.Fatalf("unexpected tag count query: %v", repo.findByTagsRequest)
	}
}

func TestQueryByFileReturnsEmptyWhenNoTagsDetected(t *testing.T) {
	calls := make([]string, 0)
	repo := &fakeMediaRepository{
		calls: &calls,
	}
	imageInference := &fakeImageInferenceClient{}
	queryService := NewQueryServiceWithAllDependencies(
		repo,
		nil,
		nil,
		imageInference,
	)

	response, err := queryService.QueryByFile(
		context.Background(),
		"query.jpg",
		"image/jpeg",
		[]byte("image bytes"),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(response.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(response.Items))
	}

	if len(calls) != 0 {
		t.Fatalf("expected no repository calls, got %v", calls)
	}
}

func TestQueryByFileReturnsInferenceError(t *testing.T) {
	repo := &fakeMediaRepository{
		calls: &[]string{},
	}
	imageInference := &fakeImageInferenceClient{
		err: errors.New("inference failed"),
	}
	queryService := NewQueryServiceWithAllDependencies(
		repo,
		nil,
		nil,
		imageInference,
	)

	_, err := queryService.QueryByFile(
		context.Background(),
		"query.jpg",
		"image/jpeg",
		[]byte("image bytes"),
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestQueryByFileRejectsNonImage(t *testing.T) {
	repo := &fakeMediaRepository{
		calls: &[]string{},
	}
	queryService := NewQueryServiceWithAllDependencies(
		repo,
		nil,
		nil,
		&fakeImageInferenceClient{},
	)

	_, err := queryService.QueryByFile(
		context.Background(),
		"query.txt",
		"text/plain",
		[]byte("not image"),
	)
	if !errors.Is(err, ErrUnsupportedQueryFile) {
		t.Fatalf("expected unsupported file error, got %v", err)
	}
}

func assertCallOrder(
	t *testing.T,
	actual []string,
	expected []string,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf(
			"expected calls %v, got %v",
			expected,
			actual,
		)
	}

	for index, expectedCall := range expected {
		if actual[index] != expectedCall {
			t.Fatalf(
				"expected call %d to be %s, got %s",
				index,
				expectedCall,
				actual[index],
			)
		}
	}
}
