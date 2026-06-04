# Media Processing

This module is Member B's GCP Cloud Run media inference service.

The final team architecture uses AWS S3 as the formal media bucket. Frontend upload goes to S3, an S3 object-created event triggers the AWS Lambda coordinator, and the Lambda coordinator calls this Cloud Run service over HTTP at `POST /infer`.

This module remains simple and testable. It does not store AWS credentials, upload artifacts to S3, write DynamoDB, or call SNS. By default it uses `fake-detector-v0`; the provided AussieEcoLense model can be enabled through local model artifacts and environment variables.

## Current Scope

Implemented in this skeleton:

- FastAPI `POST /infer` HTTP interface for the Cloud Run service
- FastAPI `GET /health` readiness endpoint for Cloud Run checks
- fake detector fallback and model version reporting
- optional provided AussieEcoLense model detector integration
- GCS model artifact loading for Cloud Run
- Cloud Run runtime dependencies for the provided model path
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

When `download_url` is not provided, `/infer` keeps the placeholder behavior: it returns `status="ready"` for supported image/video requests and uses the configured detector. Unsupported `file_type` values fail request validation.

Stage 6A adds `download_media(download_url, target_dir, filename)` in `media_downloader.py`. The helper validates that `filename` is a plain file name, writes to `/tmp` or another caller-provided temporary directory, uses a request timeout, and raises clear errors for HTTP, network, timeout, and file-write failures.

The production Cloud Run path is `POST /infer`. When `download_url` is provided, `/infer` now attempts real media download. Fake or unreachable URLs are expected to return `status="failed"` with an `error` message instead of silently returning placeholder success.

Stage 6B connects the image `download_url` path to the downloader and thumbnail helper. For image requests with `download_url`, `/infer` downloads the image into a temporary directory, generates a local JPEG thumbnail, and returns the existing `thumbnail_object_path` value. It does not upload the thumbnail to S3 yet. If download or thumbnail processing fails, `/infer` returns `status="failed"` with an `error` message instead of crashing. Image requests without `download_url` still use the placeholder behavior for Lambda coordinator testing.

Stage 6C connects the video `download_url` path to the downloader and 1 FPS frame extraction helper. For video requests with `download_url`, `/infer` downloads the video into a temporary directory, extracts sampled frames locally, runs the configured detector on each frame, and aggregates tags using the maximum count seen for each tag across sampled frames. It does not upload frames or poster images to S3 yet. If video download, frame extraction, or detection fails, `/infer` returns `status="failed"` with an `error` message instead of crashing. Video requests without `download_url` still use the placeholder behavior for Lambda coordinator testing.

Stage 8B adds optional provided model detector integration. When `USE_PROVIDED_MODEL=true` and the required model files are present, `detector.py` routes image detection through `provided_model_detector.py`. That module runs `mdv5a.pt` / MegaDetector to find animal bounding boxes, saves crop images in a temporary directory, classifies each crop with `model.pt`, and returns detections such as `{"label": "Alectura_lathami", "count": 1, "confidence": 1.0}`. The provided model package has been locally verified with the teacher's test image: MegaDetector produced `cropped_images/1-0.JPG`, and the fine-tuned classifier returned `Species: Alectura_lathami`, `Confidence: 1.0000`.

If the provided model is disabled, model files are missing, or model loading/inference fails, the service falls back to `fake-detector-v0` so local development and coordinator contract tests keep working.

Stage 8C adds Cloud Run model artifact loading from GCS. When `USE_PROVIDED_MODEL=true`, `detector.py` first asks `model_artifact_loader.py` for a usable local model directory. The loader uses a complete local `PROVIDED_MODEL_DIR` when available. If the local directory is missing or incomplete and `GCS_MODEL_BUCKET` plus `GCS_MODEL_PREFIX` are configured, it downloads `mdv5a.pt`, `model.pt`, `labels.txt`, and `config.yaml` into `MODEL_ARTIFACT_CACHE_DIR`, then points the provided detector at that cache directory. If GCS setup, download, or model loading fails, the service still falls back to `fake-detector-v0`.

Stage 8D adds the Cloud Run Python runtime dependencies needed by the provided model path: MegaDetector, PyTorch, TorchVision, and onnx2torch. These packages are installed into the container through `requirements.txt`; the model weights remain external artifacts loaded from a local directory or GCS. The heavy model packages are still lazy-imported by `provided_model_detector.py`, so startup and fake-detector fallback do not require loading the model.

## Provided Model Artifacts

The provided model artifacts are local ignored files and must not be committed to Git:

- `functions/media_processing/model_artifacts/`
- `*.pt`
- `.venv-model`
- `cropped_images`
- `mg_detections.json`

Set `PROVIDED_MODEL_DIR` to the directory that contains:

- `mdv5a.pt`
- `model.pt`
- `labels.txt`

If `PROVIDED_MODEL_DIR` is not set, the service looks in:

```bash
functions/media_processing/model_artifacts/AussieEcoLense
```

Enable the provided model route with:

```bash
USE_PROVIDED_MODEL=true
PROVIDED_MODEL_DIR=functions/media_processing/model_artifacts/AussieEcoLense
```

For Cloud Run, the model artifacts can be downloaded from GCS instead of being committed or baked into Git:

```bash
USE_PROVIDED_MODEL=true
GCS_MODEL_BUCKET=fit5225-wildlens-model-artifacts
GCS_MODEL_PREFIX=aussie-ecolense/v1
MODEL_ARTIFACT_CACHE_DIR=/tmp/aussie-ecolense/v1
```

The current GCS artifact location is:

```bash
gs://fit5225-wildlens-model-artifacts/aussie-ecolense/v1/
```

The Cloud Run service account needs read access to those objects, such as `roles/storage.objectViewer`.

The container installs the provided model runtime packages from `requirements.txt`, including `megadetector`, `onnx2torch`, `torch`, and `torchvision`. `google-cloud-storage` is included for artifact loading. The code does not pin `protobuf`; the local verified run tolerated a protobuf dependency warning, so this stays unpinned unless a Cloud Run build or runtime test proves otherwise.

## Cloud Run Deployment

This directory includes the Cloud Run `Dockerfile` and `.dockerignore`, and the service has already been deployed and tested online.

The container starts the FastAPI service with `uvicorn`:

```bash
uvicorn app:app --host 0.0.0.0 --port ${PORT:-8080}
```

The `PORT` environment variable is read by the container command, with `8080` as the local fallback.

Current endpoints:

- `GET /health` has been tested online and returns `{"status": "ok"}`.
- `POST /infer` returns inference results when no `download_url` is provided, and runs local image/video media preparation when a supported `download_url` is provided.

After deployment, the Cloud Run service URL must be shared with Member A so the AWS Lambda coordinator can call `POST /infer`.

The service still does not add or store AWS credentials.

To deploy the provided model runtime path with GCS-backed artifacts:

```powershell
gcloud run deploy wildlens-media-infer `
  --source functions/media_processing `
  --region australia-southeast1 `
  --allow-unauthenticated `
  --min-instances 0 `
  --max-instances 1 `
  --memory 4Gi `
  --cpu 2 `
  --timeout 900 `
  --concurrency 1 `
  --set-env-vars USE_PROVIDED_MODEL=true,GCS_MODEL_BUCKET=fit5225-wildlens-model-artifacts,GCS_MODEL_PREFIX=aussie-ecolense/v1,MODEL_ARTIFACT_CACHE_DIR=/tmp/aussie-ecolense/v1
```

The `4Gi` memory and `2` CPU settings are intentionally conservative for PyTorch plus MegaDetector cold starts and model inference. Reducing them is possible after a successful cloud runtime test with realistic images.

## Responsibilities

This Cloud Run service is responsible for media preparation and inference response work:

- local image thumbnail generation
- local video 1 FPS frame extraction
- fake detector fallback tagging
- optional provided model image detection and crop classification
- `tags`, `tag_counts`, `primary_species`, and `thumbnail_object_path` calculation

The AWS Lambda coordinator remains responsible for:

- reacting to S3 object-created events
- passing download/upload URLs to Cloud Run
- writing DynamoDB metadata to table `fit5225-wildlife-media-metadata`
- triggering SNS

Current Cloud Run scope still does not include S3 artifact upload, DynamoDB writes, stored AWS credentials, or SNS notifications. Real ML is optional and artifact-backed in Stage 8B/8C/8D; `fake-detector-v0` remains the fallback, not the intended provided-model path.

## Legacy Helper

`process_event()` is kept as a local test/legacy helper. It is not the final production trigger path.

The final production entrypoint is Cloud Run HTTP `POST /infer`.

## TODO For Production Integration

- Run a real Cloud Run `/infer` test with `USE_PROVIDED_MODEL=true` and a coordinator-provided image URL.
- Use coordinator-provided thumbnail upload URLs instead of fake paths.
- Upload generated thumbnails, frames, or poster images only after the team agrees on artifact paths and upload URL handling.
- Return production-ready error details for failed inference.
- Keep DynamoDB writes to `fit5225-wildlife-media-metadata` and SNS notifications in the AWS Lambda coordinator, not in this service.

Planned next stages:

- Stage 8D: Cloud Run provided model runtime support.
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
