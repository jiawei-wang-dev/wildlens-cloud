package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health returns the service status.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "wildlens-query-api",
		"status":  "ok",
	})
}
