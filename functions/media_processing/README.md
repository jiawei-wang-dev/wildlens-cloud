# Media Processing

This module is the local skeleton for Member B's GCP Cloud Run media inference service.

The final team architecture uses AWS S3 as the formal media bucket. Frontend upload goes to S3, an S3 object-created event triggers the AWS Lambda coordinator, and the Lambda coordinator calls this Cloud Run service over HTTP at `POST /infer`.

This module remains local and testable. It does not store AWS credentials, upload real thumbnails, write DynamoDB, call SNS, or run a real ML model.

## Current Scope

Implemented in this skeleton:

- FastAPI `POST /infer` HTTP interface for the future Cloud Run service
- FastAPI `GET /health` readiness endpoint for Cloud Run checks
- fake detector and model version placeholder
- image tag count aggregation and primary species selection
- video per-frame tag aggregation using the maximum species count across sampled frames
- Stage 6A `download_url` helper for downloading coordinator-provided media URLs into a temporary directory
- image thumbnail helper
- video frame extraction helper at 1 frame per second
- metadata/result models used by local tests and coordinator contract discussion
- legacy `process_event()` helper for local tests only

## Cloud Run `/infer` Contract

The AWS Lambda coordinator is expected to call `/infer` with object metadata and temporary access URLs when real integration begins.

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

Current placeholder behavior returns `status="ready"` for supported image/video requests and uses the fake detector. Unsupported `file_type` values fail request validation.

Stage 6A adds `download_media(download_url, target_dir, filename)` in `media_downloader.py` for future `download_url` integration. The helper validates that `filename` is a plain file name, writes to `/tmp` or another caller-provided temporary directory, uses a request timeout, and raises clear errors for HTTP, network, timeout, and file-write failures.

The production Cloud Run path remains `POST /infer`. This stage does not require `/infer` to successfully download real media yet, so existing placeholder Cloud Run integration can continue to work with fake URLs.

## Cloud Run Deployment Preparation

This directory now includes a Cloud Run-ready `Dockerfile` and `.dockerignore`.

The container starts the FastAPI service with `uvicorn`:

```bash
uvicorn app:app --host 0.0.0.0 --port ${PORT:-8080}
```

The `PORT` environment variable is read by the container command, with `8080` as the local fallback.

Current endpoints:

- `GET /health` returns `{"status": "ok"}`.
- `POST /infer` returns placeholder inference results matching the ML API contract.

After deployment, the Cloud Run service URL must be shared with Member A so the AWS Lambda coordinator can call `POST /infer`.

This stage does not deploy to Cloud Run and does not add credentials.

## Responsibilities

This Cloud Run service will be responsible for media inference work:

- thumbnail generation
- video 1 FPS frame extraction
- ML tagging
- `tags`, `tag_counts`, `primary_species`, and `thumbnail_object_path` calculation

The AWS Lambda coordinator remains responsible for:

- reacting to S3 object-created events
- passing download/upload URLs to Cloud Run
- writing DynamoDB metadata to table `fit5225-wildlife-media-metadata`
- triggering SNS

## Legacy Helper

`process_event()` is kept as a local test/legacy helper. It is not the final production trigger path.

The final production entrypoint is Cloud Run HTTP `POST /infer`.

## TODO For Real Integration

- Package and deploy the FastAPI service to GCP Cloud Run.
- Let the AWS Lambda coordinator call `/infer`.
- Use coordinator-provided download URLs instead of fake local paths.
- Use coordinator-provided thumbnail upload URLs instead of fake paths.
- Replace the fake detector with the selected wildlife ML model.
- Return production-ready error details for failed inference.
- Keep DynamoDB writes to `fit5225-wildlife-media-metadata` and SNS notifications in the AWS Lambda coordinator, not in this service.

Planned next stages:

- Stage 6B: real thumbnail generation/upload integration.
- Stage 6C: real video frame extraction integration.
- Stage 6D: real ML inference integration.

## Running Tests

Install dependencies for this module:

```bash
pip install -r functions/media_processing/requirements.txt
```

Run tests from the repository root:

```bash
python -m pytest functions/media_processing
```
