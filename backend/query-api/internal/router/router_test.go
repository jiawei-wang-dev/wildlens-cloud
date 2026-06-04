package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/handler"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/repository"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/router"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/service"
)

// newTestEngine creates a Gin router with temporary media data.
func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)

	files := []model.MediaFile{
		{
			FileID:              "checksum-image-001",
			OriginalFilename:    "koala.jpg",
			FileType:            "image",
			MimeType:            "image/jpeg",
			ChecksumSHA256:      "checksum-image-001",
			Bucket:              "wildlens-media",
			ObjectPath:          "media/originals/koala.jpg",
			ThumbnailObjectPath: "media/thumbnails/koala.jpg",
			FileURL:             "s3://wildlens-media/media/originals/koala.jpg",
			ThumbnailURL:        "s3://wildlens-media/media/thumbnails/koala.jpg",
			Tags:                []string{"koala", "magpie"},
			TagCounts: map[string]int{
				"koala":  3,
				"magpie": 1,
			},
			Status: "ready",
		},
		{
			FileID:           "checksum-video-001",
			OriginalFilename: "wombat.mp4",
			FileType:         "video",
			MimeType:         "video/mp4",
			ChecksumSHA256:   "checksum-video-001",
			Bucket:           "wildlens-media",
			ObjectPath:       "media/originals/wombat.mp4",
			FileURL:          "s3://wildlens-media/media/originals/wombat.mp4",
			Tags:             []string{"wombat"},
			TagCounts: map[string]int{
				"wombat": 2,
			},
			Status: "ready",
		},
	}

	repo := repository.NewMemoryRepository(files)
	queryService := service.NewQueryService(repo)
	queryHandler := handler.NewQueryHandler(queryService)

	return router.New(queryHandler)
}

// performJSONRequest sends a JSON request to the test router.
func performJSONRequest(
	engine http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(
		method,
		path,
		bytes.NewBufferString(body),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	return response
}

func TestHealth(t *testing.T) {
	engine := newTestEngine()

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload map[string]string

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", payload["status"])
	}
}

func TestFindBySpecies(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/query/species",
		`{"species":"koala"}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload struct {
		Files []model.MediaFile `json:"files"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(payload.Files))
	}

	if payload.Files[0].FileID != "checksum-image-001" {
		t.Fatalf("unexpected file ID: %s", payload.Files[0].FileID)
	}
}

func TestFindByTagCountsUsesAND(t *testing.T) {
	engine := newTestEngine()

	t.Run("returns file when all conditions match", func(t *testing.T) {
		response := performJSONRequest(
			engine,
			http.MethodPost,
			"/api/v1/query/tags",
			`{"koala":3,"magpie":1}`,
		)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", response.Code)
		}

		var payload struct {
			Files []model.MediaFile `json:"files"`
		}

		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(payload.Files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(payload.Files))
		}
	})

	t.Run("returns no file when one condition does not match", func(t *testing.T) {
		response := performJSONRequest(
			engine,
			http.MethodPost,
			"/api/v1/query/tags",
			`{"koala":3,"wombat":1}`,
		)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", response.Code)
		}

		var payload struct {
			Files []model.MediaFile `json:"files"`
		}

		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(payload.Files) != 0 {
			t.Fatalf("expected 0 files, got %d", len(payload.Files))
		}
	})
}

func TestFindOriginalByThumbnailURL(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/query/thumbnail",
		`{"thumbnail_url":"s3://wildlens-media/media/thumbnails/koala.jpg"}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload map[string]string

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "s3://wildlens-media/media/originals/koala.jpg"

	if payload["file_url"] != expected {
		t.Fatalf(
			"expected file URL %q, got %q",
			expected,
			payload["file_url"],
		)
	}
}

func TestFindBySpeciesRejectsMissingSpecies(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/query/species",
		`{}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestFindBySpeciesRejectsInvalidJSON(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/query/species",
		`{"species":`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestFindByTagCountsRejectsEmptyConditions(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/query/tags",
		`{}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestFindByTagCountsRejectsInvalidMinimumCount(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/query/tags",
		`{"koala":0}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestFindOriginalByThumbnailURLReturnsNotFound(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/query/thumbnail",
		`{"thumbnail_url":"s3://wildlens-media/media/thumbnails/missing.jpg"}`,
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}
}

func TestUpdateTagsAddsTag(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/tags/update",
		`{
			"urls": [
				"s3://wildlens-media/media/originals/koala.jpg"
			],
			"tags": [
				" reviewed "
			],
			"operation": 1
		}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload model.TagUpdateResponse

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.UpdatedCount != 1 {
		t.Fatalf(
			"expected updated_count 1, got %d",
			payload.UpdatedCount,
		)
	}

	if len(payload.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(payload.Files))
	}

	if payload.Files[0].TagCounts["reviewed"] != 1 {
		t.Fatalf(
			"expected reviewed count 1, got %d",
			payload.Files[0].TagCounts["reviewed"],
		)
	}
}

func TestUpdateTagsRemovesTagUsingThumbnailURL(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/tags/update",
		`{
			"urls": [
				"s3://wildlens-media/media/thumbnails/koala.jpg"
			],
			"tags": [
				"magpie"
			],
			"operation": 0
		}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload model.TagUpdateResponse

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.UpdatedCount != 1 {
		t.Fatalf(
			"expected updated_count 1, got %d",
			payload.UpdatedCount,
		)
	}

	if _, exists := payload.Files[0].TagCounts["magpie"]; exists {
		t.Fatal("expected magpie tag count to be removed")
	}
}

func TestUpdateTagsRejectsInvalidOperation(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/tags/update",
		`{
			"urls": [
				"s3://wildlens-media/media/originals/koala.jpg"
			],
			"tags": [
				"reviewed"
			],
			"operation": 99
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestUpdateTagsRejectsMissingOperation(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/tags/update",
		`{
			"urls": [
				"s3://wildlens-media/media/originals/koala.jpg"
			],
			"tags": [
				"reviewed"
			]
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestUpdateTagsRejectsEmptyURLs(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/tags/update",
		`{
			"urls": [],
			"tags": [
				"reviewed"
			],
			"operation": 1
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestUpdateTagsRejectsWhitespaceOnlyTags(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodPost,
		"/api/v1/tags/update",
		`{
			"urls": [
				"s3://wildlens-media/media/originals/koala.jpg"
			],
			"tags": [
				"   "
			],
			"operation": 1
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestDeleteFilesRemovesMatchedFile(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodDelete,
		"/api/v1/files",
		`{
			"file_ids": [
				"checksum-image-001"
			]
		}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload model.FileDeleteResponse

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.DeletedCount != 1 {
		t.Fatalf(
			"expected deleted_count 1, got %d",
			payload.DeletedCount,
		)
	}

	if len(payload.DeletedFileIDs) != 1 {
		t.Fatalf(
			"expected 1 deleted file ID, got %d",
			len(payload.DeletedFileIDs),
		)
	}

	if payload.DeletedFileIDs[0] != "checksum-image-001" {
		t.Fatalf(
			"unexpected deleted file ID: %s",
			payload.DeletedFileIDs[0],
		)
	}
}

func TestDeleteFilesIgnoresUnknownID(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodDelete,
		"/api/v1/files",
		`{
			"file_ids": [
				"missing-file-id"
			]
		}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload model.FileDeleteResponse

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.DeletedCount != 0 {
		t.Fatalf(
			"expected deleted_count 0, got %d",
			payload.DeletedCount,
		)
	}
}

func TestDeleteFilesRejectsEmptyIDs(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodDelete,
		"/api/v1/files",
		`{
			"file_ids": []
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestDeleteFilesRejectsWhitespaceOnlyIDs(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodDelete,
		"/api/v1/files",
		`{
			"file_ids": [
				"   "
			]
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestDeleteFilesRejectsInvalidJSON(t *testing.T) {
	engine := newTestEngine()

	response := performJSONRequest(
		engine,
		http.MethodDelete,
		"/api/v1/files",
		`{
			"file_ids":
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestListObservationsReturnsFirstPage(t *testing.T) {
	engine := newTestEngine()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/observations?limit=1",
		nil,
	)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload model.ObservationListResponse

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}

	if !payload.HasMore {
		t.Fatal("expected has_more true")
	}

	if payload.NextToken == "" {
		t.Fatal("expected non-empty next_token")
	}
}

func TestListObservationsReturnsNextPage(t *testing.T) {
	engine := newTestEngine()

	firstRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/observations?limit=1",
		nil,
	)
	firstResponse := httptest.NewRecorder()

	engine.ServeHTTP(firstResponse, firstRequest)

	var firstPayload model.ObservationListResponse

	if err := json.Unmarshal(
		firstResponse.Body.Bytes(),
		&firstPayload,
	); err != nil {
		t.Fatalf("failed to decode first response: %v", err)
	}

	nextPath := "/api/v1/observations?limit=1&next_token=" +
		firstPayload.NextToken

	secondRequest := httptest.NewRequest(
		http.MethodGet,
		nextPath,
		nil,
	)
	secondResponse := httptest.NewRecorder()

	engine.ServeHTTP(secondResponse, secondRequest)

	if secondResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			secondResponse.Code,
		)
	}

	var secondPayload model.ObservationListResponse

	if err := json.Unmarshal(
		secondResponse.Body.Bytes(),
		&secondPayload,
	); err != nil {
		t.Fatalf("failed to decode second response: %v", err)
	}

	if len(secondPayload.Items) != 1 {
		t.Fatalf(
			"expected 1 item, got %d",
			len(secondPayload.Items),
		)
	}

	if secondPayload.HasMore {
		t.Fatal("expected has_more false")
	}

	if secondPayload.NextToken != "" {
		t.Fatal("expected empty next_token")
	}
}

func TestListObservationsFiltersBySpecies(t *testing.T) {
	engine := newTestEngine()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/observations?species=koala",
		nil,
	)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload model.ObservationListResponse

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}

	if payload.Items[0].FileID != "checksum-image-001" {
		t.Fatalf(
			"unexpected file ID: %s",
			payload.Items[0].FileID,
		)
	}
}

func TestListObservationsFiltersByFileType(t *testing.T) {
	engine := newTestEngine()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/observations?file_type=video",
		nil,
	)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload model.ObservationListResponse

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}

	if payload.Items[0].FileID != "checksum-video-001" {
		t.Fatalf(
			"unexpected file ID: %s",
			payload.Items[0].FileID,
		)
	}
}

func TestListObservationsRejectsInvalidLimit(t *testing.T) {
	engine := newTestEngine()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/observations?limit=0",
		nil,
	)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestListObservationsRejectsNonIntegerLimit(t *testing.T) {
	engine := newTestEngine()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/observations?limit=abc",
		nil,
	)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestListObservationsRejectsInvalidToken(t *testing.T) {
	engine := newTestEngine()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/observations?next_token=invalid-token",
		nil,
	)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestListObservationsReturnsDisplayURLs(t *testing.T) {
	engine := newTestEngine()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/observations?species=koala",
		nil,
	)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload model.ObservationListResponse

	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}

	file := payload.Items[0]

	expectedThumbnailURL := "https://local.wildlens.test/" +
		"wildlens-media/media/thumbnails/koala.jpg"

	if file.ThumbnailDisplayURL != expectedThumbnailURL {
		t.Fatalf(
			"expected thumbnail URL %q, got %q",
			expectedThumbnailURL,
			file.ThumbnailDisplayURL,
		)
	}

	expectedDownloadURL := "https://local.wildlens.test/" +
		"wildlens-media/media/originals/koala.jpg"

	if file.FileDownloadURL != expectedDownloadURL {
		t.Fatalf(
			"expected download URL %q, got %q",
			expectedDownloadURL,
			file.FileDownloadURL,
		)
	}
}
