# Database Schema Draft and Media Metadata Contract

## Status

- Status: Draft
- Final owner: Member C
- Contributor: Member B provides the proposed media-processing output fields
- Purpose: This document is a shared contract draft between media processing and query/database modules. It is not the final database design.

## Scope

Member C is responsible for the final database schema, database choice, indexing strategy, and query implementation.

Member B is responsible for producing media metadata after file processing, such as file URLs, thumbnail URLs, tags, tag_counts, model_version, and processing status.

The fields below are proposed media metadata fields, not final database requirements.

Final field names and storage structure must be confirmed with Member C.

## Proposed Media Metadata Fields from Member B

The media-processing module may output the following metadata fields:

| Field | Description |
| --- | --- |
| file_id | Unique media record identifier. |
| owner_id | User identifier from AWS Cognito. |
| original_filename | Original uploaded filename. |
| file_type | Logical media type, such as image or video. |
| mime_type | Uploaded file MIME type. |
| checksum_sha256 | SHA-256 checksum used for deduplication. |
| size | File size in bytes. |
| storage_provider | Cloud storage provider, such as GCP. |
| bucket | Storage bucket name. |
| object_path | Object path or key inside the bucket. |
| file_url | URL for the original uploaded media. |
| thumbnail_url | URL for the generated thumbnail or preview image. |
| tags | List of detected or assigned tags. |
| tag_counts | Aggregated tag count map produced from ML inference. |
| primary_species | Main detected species, if available. |
| model_version | ML model version used for inference. |
| status | Processing status, such as pending, processing, complete, failed, or deleted. |
| created_at | Record creation timestamp. |
| updated_at | Last update timestamp. |

## Responsibility Boundary

Member B does not own the final database schema.

Member B owns the media-processing output contract.

Member C decides how these fields are stored, indexed, queried, and updated.

Member B and Member C must confirm `file_id`, `tag_counts`, `thumbnail_url`, `file_url`, `bucket`, and `object_path` before real database integration.

## Open Questions for Member C

- Should `file_id` be the same as `checksum_sha256`?
- Should `tag_counts` be stored as a map/object for AND + minimum count queries?
- Should URLs be stored as `gs://` paths, signed URL sources, or public HTTPS URLs?
- For deletion, does the query module need `bucket + object_path`, `file_url`, or both?
- For videos, should `tag_counts` be aggregated across all sampled frames at 1 FPS?
- Which database backend will be used?
