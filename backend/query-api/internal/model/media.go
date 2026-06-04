package model

// MediaFile represents an uploaded wildlife image or video.
type MediaFile struct {
	FileID              string         `json:"file_id" dynamodbav:"file_id"`
	OwnerID             string         `json:"owner_id,omitempty" dynamodbav:"owner_id,omitempty"`
	OriginalFilename    string         `json:"original_filename" dynamodbav:"original_filename"`
	FileType            string         `json:"file_type" dynamodbav:"file_type"`
	MimeType            string         `json:"mime_type,omitempty" dynamodbav:"mime_type,omitempty"`
	ChecksumSHA256      string         `json:"checksum_sha256" dynamodbav:"checksum_sha256"`
	StorageProvider     string         `json:"storage_provider,omitempty" dynamodbav:"storage_provider,omitempty"`
	Bucket              string         `json:"bucket" dynamodbav:"bucket"`
	ObjectPath          string         `json:"object_path" dynamodbav:"object_path"`
	ThumbnailObjectPath string         `json:"thumbnail_object_path,omitempty" dynamodbav:"thumbnail_object_path,omitempty"`
	FileURL             string         `json:"file_url" dynamodbav:"file_url"`
	ThumbnailURL        string         `json:"thumbnail_url,omitempty" dynamodbav:"thumbnail_url,omitempty"`
	ThumbnailDisplayURL string         `json:"thumbnail_display_url,omitempty" dynamodbav:"-"`
	FileDownloadURL     string         `json:"file_download_url,omitempty" dynamodbav:"-"`
	Tags                []string       `json:"tags" dynamodbav:"tags"`
	TagCounts           map[string]int `json:"tag_counts" dynamodbav:"tag_counts"`
	PrimarySpecies      string         `json:"primary_species,omitempty" dynamodbav:"primary_species,omitempty"`
	ModelVersion        string         `json:"model_version,omitempty" dynamodbav:"model_version,omitempty"`
	Status              string         `json:"status" dynamodbav:"status"`
	CreatedAt           string         `json:"created_at,omitempty" dynamodbav:"created_at,omitempty"`
	UpdatedAt           string         `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
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

// TagUpdateRequest adds or removes tags for multiple media files.
type TagUpdateRequest struct {
	URLs      []string `json:"urls" binding:"required"`
	Tags      []string `json:"tags" binding:"required"`
	Operation *int     `json:"operation" binding:"required"`
}

// TagUpdateResponse contains media files changed by a bulk tag update.
type TagUpdateResponse struct {
	UpdatedCount int         `json:"updated_count"`
	Files        []MediaFile `json:"files"`
}

// FileDeleteRequest deletes multiple media files by their stable IDs.
type FileDeleteRequest struct {
	FileIDs []string `json:"file_ids" binding:"required"`
}

// FileDeleteResponse contains the IDs of deleted media files.
type FileDeleteResponse struct {
	DeletedCount   int      `json:"deleted_count"`
	DeletedFileIDs []string `json:"deleted_file_ids"`
}

// ObservationListResponse contains one paginated media list response.
type ObservationListResponse struct {
	Items     []MediaFile `json:"items"`
	NextToken string      `json:"next_token"`
	HasMore   bool        `json:"has_more"`
}
