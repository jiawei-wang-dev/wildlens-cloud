# Media Processing

This module is Member B's GCP Cloud Run media inference service.

The final team architecture uses AWS S3 as the formal media bucket. Frontend upload goes to S3, an S3 object-created event triggers the AWS Lambda coordinator, and the Lambda coordinator calls this Cloud Run service over HTTP at `POST /infer`.

This module remains simple and testable. It does not store AWS credentials, upload artifacts to S3, write DynamoDB, call SNS, or run a real ML model.

## Current Scope

Implemented in this skeleton:

- FastAPI `POST /infer` HTTP interface for the Cloud Run service
- FastAPI `GET /health` readiness endpoint for Cloud Run checks
- fake detector and model version placeholder
- image tag count aggregation and primary species selection
- video per-frame tag aggregation using the maximum species count across sampled frames
- Stage 6A `download_url` helper for downloading coordinator-provided media URLs into a temporary directory
- Stage 6B image path that downloads media and generates a real local thumbnail when `download_url` is provided
- Stage 6C video path that downloads media and extracts sampled frames at 1 FPS when `download_url` is provided
- image thumbnail helper
- video frame extraction helper at 1 frame per second
- metadata/result models used by local tests and coordinator contract discussion
- legacy `process_event()` helper for local tests only

## Cloud Run `/infer` Contract

The AWS Lambda coordinator calls `/infer` with object metadata and optional temporary access URLs.

Request fields:

- `file_id`
- `bucket`
- `object_path`
- `filename`
- `file_type`
- `mime_type`
- `checksum_sha256`
- `download_url` optional
- `thumbnail_upload_url` optional

Response fields:

- `file_id`
- `tags`
- `tag_counts`
- `primary_species`
- `model_version`
- `status`
- `thumbnail_object_path`
- `error` optional

When `download_url` is not provided, `/infer` keeps the placeholder behavior: it returns `status="ready"` for supported image/video requests and uses the fake detector. Unsupported `file_type` values fail request validation.

Stage 6A adds `download_media(download_url, target_dir, filename)` in `media_downloader.py`. The helper validates that `filename` is a plain file name, writes to `/tmp` or another caller-provided temporary directory, uses a request timeout, and raises clear errors for HTTP, network, timeout, and file-write failures.

The production Cloud Run path is `POST /infer`. When `download_url` is provided, `/infer` now attempts real media download. Fake or unreachable URLs are expected to return `status="failed"` with an `error` message instead of silently returning placeholder success.

Stage 6B connects the image `download_url` path to the downloader and thumbnail helper. For image requests with `download_url`, `/infer` downloads the image into a temporary directory, generates a local JPEG thumbnail, and returns the existing `thumbnail_object_path` value. It does not upload the thumbnail to S3 yet. If download or thumbnail processing fails, `/infer` returns `status="failed"` with an `error` message instead of crashing. Image requests without `download_url` still use the placeholder behavior for Lambda coordinator testing.

Stage 6C connects the video `download_url` path to the downloader and 1 FPS frame extraction helper. For video requests with `download_url`, `/infer` downloads the video into a temporary directory, extracts sampled frames locally, runs the current fake detector on each frame, and aggregates tags using the maximum count seen for each tag across sampled frames. It does not upload frames or poster images to S3 yet. If video download, frame extraction, or detection fails, `/infer` returns `status="failed"` with an `error` message instead of crashing. Video requests without `download_url` still use the placeholder behavior for Lambda coordinator testing.

## Cloud Run Deployment

This directory includes the Cloud Run `Dockerfile` and `.dockerignore`, and the service has already been deployed and tested online.

The container starts the FastAPI service with `uvicorn`:

```bash
uvicorn app:app --host 0.0.0.0 --port ${PORT:-8080}
```

The `PORT` environment variable is read by the container command, with `8080` as the local fallback.

Current endpoints:

- `GET /health` has been tested online and returns `{"status": "ok"}`.
- `POST /infer` returns placeholder inference results when no `download_url` is provided, and runs local image/video media preparation when a supported `download_url` is provided.

After deployment, the Cloud Run service URL must be shared with Member A so the AWS Lambda coordinator can call `POST /infer`.

The service still does not add or store AWS credentials.

## Responsibilities

This Cloud Run service is responsible for media preparation and inference response work:

- local image thumbnail generation
- local video 1 FPS frame extraction
- fake detector / fallback tagging until real ML is added
- `tags`, `tag_counts`, `primary_species`, and `thumbnail_object_path` calculation

The AWS Lambda coordinator remains responsible for:

- reacting to S3 object-created events
- passing download/upload URLs to Cloud Run
- writing DynamoDB metadata to table `fit5225-wildlife-media-metadata`
- triggering SNS

Current Cloud Run scope still does not include real ML, S3 artifact upload, DynamoDB writes, stored AWS credentials, or SNS notifications.

## Legacy Helper

`process_event()` is kept as a local test/legacy helper. It is not the final production trigger path.

The final production entrypoint is Cloud Run HTTP `POST /infer`.

## TODO For Real Integration

- Replace the fake detector with the selected wildlife ML model.
- Use coordinator-provided thumbnail upload URLs instead of fake paths.
- Upload generated thumbnails, frames, or poster images only after the team agrees on artifact paths and upload URL handling.
- Return production-ready error details for failed inference.
- Keep DynamoDB writes to `fit5225-wildlife-media-metadata` and SNS notifications in the AWS Lambda coordinator, not in this service.

Planned next stages:

- Stage 6D: real ML inference integration.
- Later integration: upload generated thumbnails using coordinator-provided upload URLs.

## Running Tests

Install dependencies for this module:

```bash
pip install -r functions/media_processing/requirements.txt
```

Run tests from the repository root:

```bash
python -m pytest functions/media_processing
```
