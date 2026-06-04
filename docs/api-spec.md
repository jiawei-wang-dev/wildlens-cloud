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
