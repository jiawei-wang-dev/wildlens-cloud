package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/handler"
)

// New creates the HTTP router and registers API endpoints.
func New(queryHandler *handler.QueryHandler) *gin.Engine {
	engine := gin.Default()

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: false,
	}))

	engine.GET("/health", handler.Health)

	api := engine.Group("/api/v1")
	{
		api.POST("/query/species", queryHandler.FindBySpecies)
		api.POST("/query/tags", queryHandler.FindByTagCounts)
		api.POST("/query/thumbnail", queryHandler.FindOriginalByThumbnailURL)
		api.POST("/tags/update", queryHandler.UpdateTags)
		api.DELETE("/files", queryHandler.DeleteFiles)
		api.GET("/observations", queryHandler.ListObservations)
	}

	return engine
}
