package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/handler"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/repository"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/service"
)

func main() {
	// Temporary local data.
	// Replace this with a cloud database repository later.
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
			FileURL:             "gs://wildlens-media/media/originals/koala.jpg",
			ThumbnailURL:        "gs://wildlens-media/media/thumbnails/koala.jpg",
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
			FileURL:          "gs://wildlens-media/media/originals/wombat.mp4",
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

	router := gin.Default()

	router.GET("/health", handler.Health)

	api := router.Group("/api/v1")
	{
		api.POST("/query/species", queryHandler.FindBySpecies)
		api.POST("/query/tags", queryHandler.FindByTagCounts)
		api.POST("/query/thumbnail", queryHandler.FindOriginalByThumbnailURL)
	}

	log.Println("WildLens query API is running at http://localhost:8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
