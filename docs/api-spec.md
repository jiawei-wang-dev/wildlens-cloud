# WildLens Cloud API Specification

## Bulk Tag Update

Updates tags for multiple media files in one request.

### Endpoint

```text
POST /api/v1/tags/update
```

### Authentication

The production API requires a valid Cognito token:

```text
Authorization: Bearer <token>
```

Authentication middleware may be bypassed during local development.

### Request Body

```json
{
  "urls": [
    "s3://wildlens-media/media/originals/koala.jpg",
    "s3://wildlens-media/media/originals/wombat.mp4"
  ],
  "tags": [
    "rare",
    "reviewed"
  ],
  "operation": 1
}
```

### Request Fields

| Field       | Type           | Required | Description                           |
| ----------- | -------------- | -------- | ------------------------------------- |
| `urls`      | `List<String>` | Yes      | Original media URLs or thumbnail URLs |
| `tags`      | `List<String>` | Yes      | Tags to add or remove                 |
| `operation` | `Integer`      | Yes      | `1` adds tags; `0` removes tags       |

### Business Rules

#### Add Tags

When:

```text
operation = 1
```

the API adds each tag to every matched media file.

* Duplicate tags are ignored.
* Tags are normalised to lowercase.
* Leading and trailing spaces are removed.
* If a new tag does not already exist in `tag_counts`, its default count is set to `1`.
* Existing tag counts are preserved.

Example:

Before:

```json
{
  "tags": ["koala"],
  "tag_counts": {
    "koala": 3
  }
}
```

Request:

```json
{
  "urls": ["s3://wildlens-media/media/originals/koala.jpg"],
  "tags": ["reviewed"],
  "operation": 1
}
```

After:

```json
{
  "tags": ["koala", "reviewed"],
  "tag_counts": {
    "koala": 3,
    "reviewed": 1
  }
}
```

#### Remove Tags

When:

```text
operation = 0
```

the API removes each tag from every matched media file.

* Removing a tag that does not exist is ignored.
* Removed tags are deleted from both `tags` and `tag_counts`.

#### URL Matching

The API accepts:

* Original media URLs;
* Thumbnail URLs.

Unknown URLs are ignored. The API returns:

```json
{
  "updated_count": 0,
  "files": []
}
```

when no file matches.

### Successful Response

HTTP status:

```text
200 OK
```

Response body:

```json
{
  "updated_count": 2,
  "files": [
    {
      "file_id": "checksum-image-001",
      "file_url": "s3://wildlens-media/media/originals/koala.jpg",
      "tags": ["koala", "reviewed"],
      "tag_counts": {
        "koala": 3,
        "reviewed": 1
      }
    }
  ]
}
```

### Invalid Request

HTTP status:

```text
400 Bad Request
```

Example response:

```json
{
  "error": "operation must be 0 (remove) or 1 (add)"
}
```

Invalid cases include:

* `urls` is empty;
* `tags` is empty;
* Tags contain only whitespace;
* `operation` is not `0` or `1`;
* JSON body is invalid.

### Notes

Adding tags may trigger SNS notifications in a later integration step.

The first local implementation uses `MemoryRepository`. DynamoDB persistence is added separately.

## Bulk Media Deletion

Deletes multiple media files by their stable file IDs.

### Endpoint

```text
DELETE /api/v1/files
```

### Authentication

The production API requires a valid Cognito token:

```text
Authorization: Bearer <token>
```

Authentication middleware may be bypassed during local development.

### Request Body

```json
{
  "file_ids": [
    "checksum-image-001",
    "checksum-video-001"
  ]
}
```

### Request Fields

| Field      | Type           | Required | Description                                                               |
| ---------- | -------------- | -------- | ------------------------------------------------------------------------- |
| `file_ids` | `List<String>` | Yes      | Stable media record IDs. Each `file_id` is based on the SHA-256 checksum. |

### Business Rules

* Duplicate file IDs are ignored.
* Leading and trailing spaces are removed.
* Unknown file IDs are ignored.
* Empty file IDs are ignored.
* The backend uses `file_id` rather than URLs because URLs may change or expire.
* The production implementation removes the original S3 object, the thumbnail object when present, and the DynamoDB metadata record.
* The local memory implementation removes only the in-memory metadata record.

### Successful Response

HTTP status:

```text
200 OK
```

Response body:

```json
{
  "deleted_count": 2,
  "deleted_file_ids": [
    "checksum-image-001",
    "checksum-video-001"
  ]
}
```

### No Matching Records

Unknown file IDs do not cause an error.

Example response:

```json
{
  "deleted_count": 0,
  "deleted_file_ids": []
}
```

### Invalid Request

HTTP status:

```text
400 Bad Request
```

Example response:

```json
{
  "error": "at least one file_id is required"
}
```

Invalid cases include:

* `file_ids` is empty;
* `file_ids` contains only whitespace;
* JSON body is invalid.

### Notes

The first local implementation uses `MemoryRepository`.

DynamoDB `DeleteItem` and S3 object deletion are added separately.

## Observation List

Returns a paginated list of uploaded wildlife media records.

### Endpoint

```text
GET /api/v1/observations
```

### Query Parameters

| Parameter    | Type    | Required | Description                                                           |
| ------------ | ------- | -------- | --------------------------------------------------------------------- |
| `limit`      | Integer | No       | Number of records returned per request. Default: `10`. Maximum: `50`. |
| `next_token` | String  | No       | Opaque token returned by the previous request.                        |
| `species`    | String  | No       | Filters files containing the species tag.                             |
| `file_type`  | String  | No       | Filters records by `image` or `video`.                                |
| `status`     | String  | No       | Filters records by processing status, such as `ready`.                |

### Example Request

```text
GET /api/v1/observations?limit=10&species=koala&status=ready
```

### Successful Response

```json
{
  "items": [
    {
      "file_id": "checksum-image-001",
      "original_filename": "koala.jpg",
      "file_type": "image",
      "primary_species": "koala",
      "tags": [
        "koala",
        "magpie"
      ],
      "tag_counts": {
        "koala": 3,
        "magpie": 1
      },
      "status": "ready"
    }
  ],
  "next_token": "",
  "has_more": false
}
```

### Pagination Rules

* The first request does not include `next_token`.
* When another page exists, the backend returns a non-empty `next_token`.
* The frontend must pass the returned token back unchanged.
* When there is no next page, `next_token` is an empty string and `has_more` is `false`.
* The token is opaque. The frontend must not decode or modify it.

### Invalid Request

HTTP status:

```text
400 Bad Request
```

Invalid cases include:

* `limit` is not an integer;
* `limit` is less than `1`;
* `limit` is greater than `50`;
* `next_token` is invalid.

### Notes

The first implementation may use DynamoDB `Scan` followed by filtering and pagination in the Go service.

The frontend contract remains unchanged if the backend later adds a DynamoDB Global Secondary Index or native DynamoDB cursor pagination.
