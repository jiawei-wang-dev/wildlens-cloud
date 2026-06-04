package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/repository"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/service"
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
	var request model.TagCountQueryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
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

// DeleteFiles removes multiple media files by their stable IDs.
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
		request.FileIDs,
	)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, service.ErrFileIDsRequired) {
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
