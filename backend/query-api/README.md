# WildLens Query API

Go API for querying WildLens media metadata.

## Local Memory Mode

```bash
cd backend/query-api
APP_REPOSITORY=memory go run ./cmd/server
```

The local memory mode does not require AWS credentials.

## Local Tests

```bash
cd backend/query-api
go fmt ./...
go mod tidy
go test -v ./...
```

## API Contracts

The canonical observation list route is:

```text
GET /api/v1/observations
```

Each item keeps these frontend fields:

- `file_url`: pass back to `POST /api/v1/tags/update` and `DELETE /api/v1/files`.
- `thumbnail_display_url`: render thumbnails in the table.
- `file_download_url`: download the original image or video.

Advanced tag count query accepts the official JSON map format:

```http
POST /api/v1/query/tags
Content-Type: application/json
```

```json
{
  "koala": 3,
  "wombat": 2
}
```

The query uses AND logic. The compatibility wrapper remains supported:

```json
{
  "tag_counts": {
    "koala": 3,
    "wombat": 2
  }
}
```

Thumbnail reverse lookup canonical route:

```http
POST /api/v1/query/thumbnail
Content-Type: application/json
```

```json
{
  "thumbnail_url": "https://example.com/media/thumbnails/checksum.jpg"
}
```

Compatibility alias:

```text
GET /api/v1/observations/lookup?thumbnail_url=<URL_STRING>
```

Both return:

```json
{
  "file_url": "https://example.com/media/originals/koala.jpg"
}
```

Temporary image search canonical route:

```http
POST /api/v1/query/file
Content-Type: multipart/form-data
```

Compatibility alias:

```http
POST /api/v1/observations/search-by-file
Content-Type: multipart/form-data
```

Use multipart field `file`. The API accepts `image/jpeg` and `image/png` up to 10 MiB. Temporary query files are only sent to the stateless inference service configured by `TEMP_QUERY_INFER_URL`; they are not uploaded to S3, not written to DynamoDB, and not retained after the request.

Expected stateless inference contract:

```http
POST ${TEMP_QUERY_INFER_URL}
Content-Type: multipart/form-data
```

Multipart field:

```text
file
```

Expected response:

```json
{
  "tags": ["koala", "magpie"],
  "tag_counts": {
    "koala": 1,
    "magpie": 1
  },
  "primary_species": "koala",
  "model_version": "provided-aussie-ecolense-v1"
}
```

Query API response:

```json
{
  "detected_tags": ["koala", "magpie"],
  "items": []
}
```

## Cloud Run Deployment

Run from the repository root:

```bash
chmod +x backend/query-api/scripts/deploy-cloud-run.sh
./backend/query-api/scripts/deploy-cloud-run.sh
```

The deploy script silently prompts for temporary AWS Academy credentials and writes them to Google Cloud Secret Manager. It does not print the credential values.

Before running the script, make sure Google Cloud CLI is installed, you are logged in, and a project is selected:

```bash
gcloud auth login
gcloud config set project <PROJECT_ID>
```

## Refresh AWS Academy Credentials

AWS Academy Lab credentials change after the lab restarts. Refresh the Cloud Run secrets without changing code:

```bash
chmod +x backend/query-api/scripts/refresh-aws-secrets.sh
./backend/query-api/scripts/refresh-aws-secrets.sh
```

The refresh script adds new Secret Manager versions and updates Cloud Run to use the latest versions.

## Security

- Do not commit real AWS credentials.
- Do not write credentials into `.env`.
- Do not paste credentials into chat, group chat, or Codex.
- After AWS Academy credentials expire, run the refresh script.
