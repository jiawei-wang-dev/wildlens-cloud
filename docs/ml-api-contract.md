# ML API Contract

## Status

- Status: Confirmed v1 contract
- Caller: AWS Lambda coordinator
- Provider: GCP Cloud Run
- Endpoint: `POST /infer`
- Purpose: media processing and ML inference for uploaded image/video objects

The Cloud Run service is called after an AWS S3 object-created event has triggered the AWS Lambda coordinator.

## Responsibility Boundary

The GCP Cloud Run media inference service:

- performs thumbnail, video frame extraction, and ML tagging work
- returns inference results to the AWS Lambda coordinator
- does not write DynamoDB directly
- does not store AWS credentials

The AWS Lambda coordinator:

- receives S3 object-created events
- calls Cloud Run `POST /infer`
- passes short-lived S3 download URLs when real integration is implemented
- passes thumbnail upload URLs when thumbnail upload is implemented
- writes DynamoDB table `fit5225-wildlife-media-metadata`
- may trigger SNS notifications

## Request Body

```json
{
  "file_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "bucket": "fit5225-wildlife-media",
  "object_path": "incoming/user-1/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/koala.jpg",
  "filename": "koala.jpg",
  "file_type": "image",
  "mime_type": "image/jpeg",
  "checksum_sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "download_url": "https://example.test/short-lived-download-url",
  "thumbnail_upload_url": "https://example.test/short-lived-thumbnail-upload-url"
}
```

| Field | Required | Description |
| --- | --- | --- |
| file_id | Yes | Media identifier. Must equal `checksum_sha256`. |
| bucket | Yes | Official AWS S3 media bucket. |
| object_path | Yes | S3 object key for the uploaded original media. |
| filename | Yes | Original safe filename. |
| file_type | Yes | `image` or `video`. |
| mime_type | Yes | Uploaded file MIME type. |
| checksum_sha256 | Yes | 64-character SHA-256 checksum. |
| download_url | No | Short-lived S3 download URL supplied by the AWS Lambda coordinator for real integration. |
| thumbnail_upload_url | No | Short-lived upload URL for generated thumbnail or preview artifacts. |

## Response Body

```json
{
  "file_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "tags": ["koala"],
  "tag_counts": {
    "koala": 1
  },
  "primary_species": "koala",
  "model_version": "fake-detector-v0",
  "status": "ready",
  "thumbnail_object_path": "media/thumbnails/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg",
  "error": null
}
```

| Field | Description |
| --- | --- |
| file_id | Same identifier received in the request. |
| tags | Deduplicated list of detected tags. |
| tag_counts | Map/object of detected species to numeric counts. |
| primary_species | Main detected species, if available. |
| model_version | ML model version used for inference. |
| status | `ready` or `failed`. |
| thumbnail_object_path | S3 object key for generated thumbnail or preview image, if available. |
| error | Optional failure detail when `status="failed"`. |

## Aggregation Rules

For images, `tag_counts` is calculated by summing detections for matching species within the image.

For videos, Cloud Run samples frames at 1 FPS. `tags` is the union of sampled frame tags. `tag_counts` uses the maximum detected count per species across sampled frames, not the sum across all sampled frames.

## Current Skeleton Behavior

The current local skeleton exposes `POST /infer` with FastAPI and returns placeholder inference results. It does not download real S3 files, upload thumbnails, run a real ML model, write DynamoDB, store AWS credentials, deploy Cloud Run, or trigger SNS.
