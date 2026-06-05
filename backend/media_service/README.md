# Media Service

This is the Stage 2 local skeleton for Member B's upload and deduplication API.

The service is intentionally local and testable. It does not connect to real AWS, GCP, DynamoDB, signed URL services, or credentials.

## Scope

- `POST /media/upload-url` validates upload metadata before a client uploads media.
- `file_id` is always `checksum_sha256`.
- `checksum_sha256` must be a 64-character SHA-256 hex string. Uppercase and lowercase hex are accepted.
- `filename` must be a plain filename, not a path. It cannot be blank, `.`, `..`, or include `/` or `\`.
- Checksum deduplication uses only `checksum_sha256`, never `filename`.
- New uploads use the object path format `incoming/{owner_id}/{checksum_sha256}/{filename}`.
- Duplicate uploads return existing file and thumbnail URLs from the fake database client.
- Non-duplicate uploads return a fake upload URL from an abstract storage client.
- The current upload URL is a placeholder. The future implementation will generate AWS S3 presigned upload URLs.

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

The final media bucket is AWS S3. The team bucket name is currently `fit5225-wildlife-media`, but Python code should read the bucket from `MEDIA_BUCKET` instead of hard-coding it.

Future upload URL implementation:

- Generate an AWS S3 presigned upload URL.
- Required environment variables: `AWS_REGION`, `MEDIA_BUCKET`.
- Frontend uploads directly to S3 with that URL.
- S3 object-created events trigger the AWS Lambda coordinator.
- The Lambda coordinator calls the GCP Cloud Run media inference service at `/infer`.
- The Lambda coordinator writes DynamoDB metadata to table `fit5225-wildlife-media-metadata` and triggers SNS. This upload API skeleton does not write DynamoDB directly.
- Future deduplication lookup will use DynamoDB `GetItem` with `file_id = checksum_sha256` against table `fit5225-wildlife-media-metadata`.

## Running Tests

Install dependencies for this module:

```bash
pip install -r backend/media_service/requirements.txt
```

Run tests from the repository root:

```bash
python -m pytest backend/media_service
```
