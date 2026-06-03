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
