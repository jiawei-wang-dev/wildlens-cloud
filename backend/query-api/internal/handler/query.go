package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/repository"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/service"
)

const (
	maxQueryFileBytes   = 10 * 1024 * 1024
	maxQueryRequestBody = maxQueryFileBytes + 1024*1024
)

// QueryHandler handles media query requests.
type QueryHandler struct {
	service *service.QueryService
}

func NewQueryHandler(queryService *service.QueryService) *QueryHandler {
	return &QueryHandler{service: queryService}
}

// FindBySpecies returns files containing at least one requested species.
func (h *QueryHandler) FindBySpecies(c *gin.Context) {
	var request model.SpeciesQueryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body or missing species",
		})
		return
	}

	files, err := h.service.FindBySpecies(
		c.Request.Context(),
		request.Species,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// FindByTagCounts applies logical AND between minimum-count conditions.
func (h *QueryHandler) FindByTagCounts(c *gin.Context) {
	request, err := parseTagCountQueryRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
		})
		return
	}

	files, err := h.service.FindByTagCounts(
		c.Request.Context(),
		request,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

func parseTagCountQueryRequest(
	c *gin.Context,
) (model.TagCountQueryRequest, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return model.TagCountQueryRequest{}, nil
	}

	if tagCountsRaw, exists := raw["tag_counts"]; exists {
		var wrapped model.TagCountQueryRequest

		if err := json.Unmarshal(tagCountsRaw, &wrapped); err != nil {
			return nil, err
		}

		return wrapped, nil
	}

	var request model.TagCountQueryRequest

	if err := json.Unmarshal(body, &request); err != nil {
		return nil, err
	}

	return request, nil
}

// FindOriginalByThumbnailURL maps a thumbnail URL to its original file URL.
func (h *QueryHandler) FindOriginalByThumbnailURL(c *gin.Context) {
	var request model.ThumbnailQueryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body or missing thumbnail_url",
		})
		return
	}

	fileURL, err := h.service.FindOriginalByThumbnailURL(
		c.Request.Context(),
		request.ThumbnailURL,
	)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, repository.ErrMediaNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_url": fileURL,
	})
}

// LookupOriginalByThumbnailURL maps a thumbnail URL query parameter to the original URL.
func (h *QueryHandler) LookupOriginalByThumbnailURL(c *gin.Context) {
	fileURL, err := h.service.FindOriginalByThumbnailURL(
		c.Request.Context(),
		c.Query("thumbnail_url"),
	)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, service.ErrThumbnailURLRequired) {
			status = http.StatusBadRequest
		}

		if errors.Is(err, repository.ErrMediaNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_url": fileURL,
	})
}

// QueryByFile searches observations using stateless inference on a temporary image.
func (h *QueryHandler) QueryByFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxQueryRequestBody,
	)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing multipart file field",
		})
		return
	}

	if fileHeader.Size > maxQueryFileBytes {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query file exceeds 10 MiB limit",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "cannot read query file",
		})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxQueryFileBytes+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "cannot read query file",
		})
		return
	}

	if len(data) > maxQueryFileBytes {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query file exceeds 10 MiB limit",
		})
		return
	}

	contentType := strings.TrimSpace(
		fileHeader.Header.Get("Content-Type"),
	)
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	response, err := h.service.QueryByFile(
		c.Request.Context(),
		fileHeader.Filename,
		contentType,
		data,
	)
	if err != nil {
		status := http.StatusBadGateway

		if errors.Is(err, service.ErrQueryFileRequired) ||
			errors.Is(err, service.ErrUnsupportedQueryFile) {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateTags adds or removes tags for multiple media files.
func (h *QueryHandler) UpdateTags(c *gin.Context) {
	var request model.TagUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body or missing required fields",
		})
		return
	}

	files, err := h.service.UpdateTags(
		c.Request.Context(),
		request.URLs,
		request.Tags,
		request.Operation,
	)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, service.ErrURLsRequired) ||
			errors.Is(err, service.ErrTagsRequired) ||
			errors.Is(err, service.ErrTagOperationRequired) ||
			errors.Is(err, repository.ErrInvalidTagOperation) {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.TagUpdateResponse{
		UpdatedCount: len(files),
		Files:        files,
	})
}

// DeleteFiles removes multiple media files by their stable URLs.
func (h *QueryHandler) DeleteFiles(c *gin.Context) {
	var request model.FileDeleteRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body or missing required fields",
		})
		return
	}

	deletedFileIDs, err := h.service.DeleteFiles(
		c.Request.Context(),
		request.URLs,
	)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, service.ErrURLsRequired) {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.FileDeleteResponse{
		DeletedCount:   len(deletedFileIDs),
		DeletedFileIDs: deletedFileIDs,
	})
}

// ListObservations returns filtered and paginated media records.
func (h *QueryHandler) ListObservations(c *gin.Context) {
	limit := repository.DefaultObservationLimit
	limitValue := strings.TrimSpace(c.Query("limit"))

	if limitValue != "" {
		parsedLimit, err := strconv.Atoi(limitValue)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be an integer",
			})
			return
		}

		limit = parsedLimit
	}

	page, err := h.service.ListObservations(
		c.Request.Context(),
		limit,
		c.Query("next_token"),
		c.Query("species"),
		c.QueryArray("tag"),
		c.Query("file_type"),
		c.Query("status"),
	)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, service.ErrInvalidObservationLimit) ||
			errors.Is(err, repository.ErrInvalidNextToken) {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.ObservationListResponse{
		Items:     page.Items,
		NextToken: page.NextToken,
		HasMore:   page.HasMore,
	})
}
