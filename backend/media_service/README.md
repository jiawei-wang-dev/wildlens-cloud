# Media Service

This is the Stage 2 local skeleton for Member B's upload and deduplication API.

The service is intentionally local and testable. It does not connect to real AWS, GCP, Firestore, signed URL services, or credentials.

## Scope

- `POST /media/upload-url` validates upload metadata before a client uploads media.
- `file_id` is always `checksum_sha256`.
- `checksum_sha256` must be a 64-character SHA-256 hex string. Uppercase and lowercase hex are accepted.
- `filename` must be a plain filename, not a path. It cannot be blank, `.`, `..`, or include `/` or `\`.
- Checksum deduplication uses only `checksum_sha256`, never `filename`.
- New uploads use the object path format `incoming/{owner_id}/{checksum_sha256}/{filename}`.
- Duplicate uploads return existing file and thumbnail URLs from the fake database client.
- Non-duplicate uploads return a fake upload URL from an abstract storage client.

## Request Body

```json
{
  "filename": "koala.jpg",
  "content_type": "image/jpeg",
  "size": 12345,
  "checksum_sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "file_type": "image"
}
```

`file_type` must be either `image` or `video`.

## Future Integration Notes

A provided the current Cognito configuration:

- Region: `us-east-1`
- User Pool ID: `us-east-1_zHpMn5rX5`
- Client ID: `4vaujh7rjc3as3q38pqlva6iet`

These values should be supplied later as environment variables. Real Cognito JWT validation is not implemented in this skeleton.

Firestore is the likely future database backend for checksum deduplication, but this service currently uses a fake in-memory client only.

The final storage target is still pending team confirmation. The planned direction is GCP Cloud Storage, but `storage_client.py` remains provider-abstract and does not implement real signed URLs. `MEDIA_BUCKET` and `MEDIA_STORAGE_PROVIDER` are reserved for future configuration; the current fallback bucket is only a placeholder. A's S3 bucket is not treated as the formal media processing bucket unless the team changes the architecture.

## Running Tests

Install dependencies for this module:

```bash
pip install -r backend/media_service/requirements.txt
```

Run tests from the repository root:

```bash
python -m pytest backend/media_service
```
