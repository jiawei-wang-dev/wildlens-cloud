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

Deletes multiple media metadata records by their stable original or thumbnail URLs.

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
  "urls": [
    "s3://wildlens-media/media/originals/koala.jpg",
    "s3://wildlens-media/media/originals/wombat.mp4"
  ]
}
```

### Request Fields

| Field  | Type           | Required | Description                                        |
| ------ | -------------- | -------- | -------------------------------------------------- |
| `urls` | `List<String>` | Yes      | Stable original file URLs or thumbnail file URLs. |

### Business Rules

* The external API accepts `urls`, not `file_ids`.
* Leading and trailing URL spaces are removed.
* Duplicate URLs are ignored.
* Unknown URLs are ignored without error.
* Empty URLs are ignored.
* If no valid URL remains after cleanup, the API returns `400 Bad Request`.
* The backend still deletes DynamoDB metadata internally by using the matched metadata record's `file_id`.
* The current version deletes metadata only. S3 original files, videos, and thumbnails will be deleted in a later integration.

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

Unknown URLs do not cause an error.

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
  "error": "at least one URL is required"
}
```

Invalid cases include:

* `urls` is empty;
* `urls` contains only whitespace;
* `urls` is missing;
* JSON body is invalid.

### Notes

The first local implementation uses `MemoryRepository`.

DynamoDB metadata deletion uses `DeleteItem`. S3 object deletion is added separately.

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
| `tag`        | String  | No       | Filters records containing the tag. Repeat the parameter to require multiple tags using AND logic. |
| `file_type`  | String  | No       | Filters records by `image` or `video`.                                |
| `status`     | String  | No       | Filters records by processing status, such as `ready`.                |

### Example Request

```text
GET /api/v1/observations?limit=10&species=koala&status=ready
```

```text
GET /api/v1/observations?species=koala&tag=wild&tag=cute&limit=10
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
* The frontend should pass the same filters again when requesting the next page.
* When there is no next page, `next_token` is an empty string and `has_more` is `false`.
* The token is opaque. The frontend must not decode or modify it.

### Filtering Rules

* Multiple `tag` parameters use AND logic.
* Empty `tag` values are ignored.

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

## Temporary Media Display URLs

Private S3 object paths cannot be rendered directly by a browser.

The observation list endpoint returns temporary HTTPS URLs for frontend display and download.

### Response Fields

| Field                   | Type   | Description                                                     |
| ----------------------- | ------ | --------------------------------------------------------------- |
| `thumbnail_display_url` | String | Temporary HTTPS URL used by the frontend to render a thumbnail. |
| `file_download_url`     | String | Temporary HTTPS URL used to download the original media file.   |

Example response item:

```json
{
  "file_id": "checksum-image-001",
  "bucket": "wildlens-media",
  "object_path": "media/originals/koala.jpg",
  "thumbnail_object_path": "media/thumbnails/koala.jpg",
  "thumbnail_display_url": "https://temporary-signed-url",
  "file_download_url": "https://temporary-signed-url"
}
```

### Frontend Usage

The frontend thumbnail component should use:

```html
<img :src="item.thumbnail_display_url" />
```

The frontend should not attempt to render:

* `thumbnail_object_path`;
* `object_path`;
* `s3://...` URLs.

### Storage Rules

The database stores stable S3 fields:

* `bucket`;
* `object_path`;
* `thumbnail_object_path`.

Temporary display URLs are generated when the API returns data.

Temporary URLs are not stored permanently in DynamoDB because they expire.

### Local Development

Local memory mode returns predictable placeholder HTTPS URLs.

AWS deployment mode returns real S3 Presigned GET URLs.
