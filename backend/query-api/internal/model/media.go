package model

// MediaFile represents an uploaded wildlife image or video.
type MediaFile struct {
	FileID              string         `json:"file_id"`
	OwnerID             string         `json:"owner_id,omitempty"`
	OriginalFilename    string         `json:"original_filename"`
	FileType            string         `json:"file_type"`
	MimeType            string         `json:"mime_type,omitempty"`
	ChecksumSHA256      string         `json:"checksum_sha256"`
	Bucket              string         `json:"bucket"`
	ObjectPath          string         `json:"object_path"`
	ThumbnailObjectPath string         `json:"thumbnail_object_path,omitempty"`
	FileURL             string         `json:"file_url"`
	ThumbnailURL        string         `json:"thumbnail_url,omitempty"`
	Tags                []string       `json:"tags"`
	TagCounts           map[string]int `json:"tag_counts"`
	PrimarySpecies      string         `json:"primary_species,omitempty"`
	ModelVersion        string         `json:"model_version,omitempty"`
	Status              string         `json:"status"`
	CreatedAt           string         `json:"created_at,omitempty"`
	UpdatedAt           string         `json:"updated_at,omitempty"`
}

// SpeciesQueryRequest searches for files containing a species.
type SpeciesQueryRequest struct {
	Species string `json:"species" binding:"required"`
}

// TagCountQueryRequest searches using minimum tag counts.
// Example: {"koala": 3, "magpie": 1}
type TagCountQueryRequest map[string]int

// ThumbnailQueryRequest maps a thumbnail URL to its original file URL.
type ThumbnailQueryRequest struct {
	ThumbnailURL string `json:"thumbnail_url" binding:"required"`
}
