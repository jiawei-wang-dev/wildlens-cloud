# Database Schema and Media Metadata Contract

## Table

- Status: Confirmed v1 contract
- Database owner: Member C
- Media metadata producer: AWS Lambda coordinator, using results from Member B's Cloud Run inference service
- Database backend: AWS DynamoDB

This document records the agreed v1 metadata table contract for the hybrid AWS/GCP media pipeline.

## Hybrid Architecture Boundary

- AWS S3 is the official media bucket.
- S3 object-created events trigger the AWS Lambda coordinator.
- The AWS Lambda coordinator calls the GCP Cloud Run media inference service at `POST /infer`.
- The GCP Cloud Run service returns inference results such as `tags`, `tag_counts`, `primary_species`, `thumbnail_object_path`, and `status`.
- The AWS Lambda coordinator writes media metadata to DynamoDB and may trigger SNS notifications.
- The GCP Cloud Run service does not write DynamoDB directly and does not store AWS credentials.
- Member C's Go service reads, modifies, and deletes items from the same DynamoDB table.

## DynamoDB Table

| Property | Value |
| --- | --- |
| Table name | `fit5225-wildlife-media-metadata` |
| Partition key | `file_id` |
| Partition key type | String |
| Sort key | None for v1 |
| GSI | None for v1 |
| Item model | One item per original image/video |

`file_id` is equal to `checksum_sha256`.

Video frame results are aggregated back into the original video item. The table should not store one item per sampled frame in v1.

Deletion in v1 removes the S3 storage objects and the DynamoDB item. Do not use `deleted` as an active v1 status.

## Media Metadata Fields

| Field | Type / Shape | Description |
| --- | --- | --- |
| file_id | String | Primary identifier and DynamoDB partition key. Equal to `checksum_sha256`. |
| owner_id | String | User identifier from AWS Cognito. |
| original_filename | String | Original uploaded filename. |
| file_type | String | Logical media type, `image` or `video`. |
| mime_type | String | Uploaded file MIME type. |
| checksum_sha256 | String | 64-character SHA-256 checksum used for deduplication. |
| size | Number | File size in bytes. |
| storage_provider | String | Storage provider for the official media object, expected to be `s3` for v1. |
| bucket | String | S3 bucket name. |
| object_path | String | Object key inside the bucket. |
| thumbnail_object_path | String or null | Thumbnail or preview object key inside the bucket. |
| file_url | String | URL or URI for the original uploaded media. |
| thumbnail_url | String or null | URL or URI for the generated thumbnail or preview image. |
| tags | List<String> | Deduplicated list of detected tags. |
| tag_counts | Map<String, Number> | Aggregated tag count map produced from ML inference. |
| primary_species | String or null | Main detected species, if available. |
| model_version | String or null | ML model version used for inference. |
| status | String | One of `pending`, `processing`, `ready`, or `failed`. |
| created_at | String | Record creation timestamp. |
| updated_at | String | Last update timestamp. |

## Video Aggregation

For videos, the Cloud Run inference service samples frames at 1 FPS. The response `tags` field is the union of sampled frame tags. The response `tag_counts` field uses the maximum detected count per species across sampled frames, not the sum across all frames.

The SHA-256 checksum is used as the unique identifier of an uploaded file and helps prevent duplicate records.

AWS Lambda coordinator:

- handles S3 object-created events
- calls GCP Cloud Run `POST /infer`
- writes items to DynamoDB table `fit5225-wildlife-media-metadata`
- may trigger SNS notifications

GCP Cloud Run media inference service:

- performs media processing and ML inference
- returns inference results to the AWS Lambda coordinator
- does not write DynamoDB directly
- does not store AWS credentials

Member C Go service:

- reads media metadata from `fit5225-wildlife-media-metadata`
- modifies supported fields as required by query/update workflows
- deletes the DynamoDB item when deleting media metadata
## Item Meaning

Each DynamoDB item represents one original uploaded image or video.

Video frames are temporary processing results. They should be aggregated before writing the final metadata record. Each extracted frame should not create a separate DynamoDB item.

## Media Item Example

```json
{
  "file_id": "sha256-checksum",
  "owner_id": "cognito-user-id",
  "original_filename": "koala.jpg",
  "file_type": "image",
  "mime_type": "image/jpeg",
  "checksum_sha256": "sha256-checksum",

  "storage_provider": "aws",
  "bucket": "fit5225-wildlife-media",
  "object_path": "media/originals/koala.jpg",
  "thumbnail_object_path": "media/thumbnails/koala.jpg",

  "file_url": "s3://fit5225-wildlife-media/media/originals/koala.jpg",
  "thumbnail_url": "s3://fit5225-wildlife-media/media/thumbnails/koala.jpg",

  "tags": ["koala", "magpie"],
  "tag_counts": {
    "koala": 3,
    "magpie": 1
  },

  "primary_species": "koala",
  "model_version": "wildlife-v1",
  "status": "ready",
  "created_at": "2026-06-03T10:00:00Z",
  "updated_at": "2026-06-03T10:00:00Z"
}
```

## Field Types

| Field                   | DynamoDB Type       | Description                                   |
| ----------------------- | ------------------- | --------------------------------------------- |
| `file_id`               | String              | SHA-256 checksum and partition key            |
| `owner_id`              | String              | Cognito user ID                               |
| `original_filename`     | String              | Original uploaded file name                   |
| `file_type`             | String              | `image` or `video`                            |
| `mime_type`             | String              | Example: `image/jpeg`                         |
| `checksum_sha256`       | String              | File checksum                                 |
| `storage_provider`      | String              | `aws`                                         |
| `bucket`                | String              | S3 bucket name                                |
| `object_path`           | String              | Original S3 object path                       |
| `thumbnail_object_path` | String              | Thumbnail S3 object path for images           |
| `file_url`              | String              | Original image or video URL                   |
| `thumbnail_url`         | String              | Thumbnail URL for images                      |
| `tags`                  | List<String>        | Deduplicated species list                     |
| `tag_counts`            | Map<String, Number> | Species count map                             |
| `primary_species`       | String              | Main detected species                         |
| `model_version`         | String              | ML model configuration version                |
| `status`                | String              | `pending`, `processing`, `ready`, or `failed` |
| `created_at`            | String              | ISO 8601 timestamp                            |
| `updated_at`            | String              | ISO 8601 timestamp                            |

## Deletion Behaviour

The first version does not use a `deleted` status.

When a user deletes files, the system should remove:

1. Original image or video objects from storage.
2. Thumbnail objects from storage.
3. Corresponding DynamoDB items.

## Indexes

The first version does not require a Global Secondary Index.

The initial query service may use DynamoDB Scan and apply dynamic tag filters in the Go service.

A future index may be added for user upload history:

```text
Partition key: owner_id
Sort key: created_at
```
