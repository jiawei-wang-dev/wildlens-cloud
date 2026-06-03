# WildLens Cloud DynamoDB Media Schema

## Table

Table name:

```text
fit5225-wildlife-media-metadata
```

Partition key:

```text
file_id
```

Partition key type:

```text
String
```

The first version does not require a sort key.

## File ID

```text
file_id = checksum_sha256
```

The SHA-256 checksum is used as the unique identifier of an uploaded file and helps prevent duplicate records.

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
