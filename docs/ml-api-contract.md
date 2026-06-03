# WildLens Cloud ML Inference API Contract

## Purpose

The GCP Cloud Run ML service processes wildlife images and videos.

It receives a short-lived URL for an uploaded S3 object, performs media processing and ML inference, and returns detected tags and counts to the AWS Lambda coordinator.

## Endpoint

```text
POST /infer
```

## Image Request

```json
{
  "file_id": "sha256-checksum",
  "file_type": "image",
  "source_url": "https://temporary-presigned-s3-get-url",
  "thumbnail_upload_url": "https://temporary-presigned-s3-put-url",
  "model_version": "wildlife-v1"
}
```

## Video Request

```json
{
  "file_id": "sha256-checksum",
  "file_type": "video",
  "source_url": "https://temporary-presigned-s3-get-url",
  "model_version": "wildlife-v1"
}
```

## Successful Response

```json
{
  "tags": ["koala", "magpie"],
  "tag_counts": {
    "koala": 3,
    "magpie": 1
  },
  "primary_species": "koala",
  "thumbnail_generated": true,
  "model_version": "wildlife-v1",
  "status": "ready"
}
```

## Error Response

```json
{
  "error": "failed to process media",
  "status": "failed"
}
```

## Processing Rules

### Images

* Download the source image using the temporary URL.
* Generate a compressed thumbnail while preserving the aspect ratio.
* Upload the thumbnail through the provided temporary upload URL.
* Run ML inference and return detected tags and counts.

### Videos

* Download the source video using the temporary URL.
* Extract one frame per second.
* Run ML inference on extracted frames.
* Aggregate detected tags and counts.
* Do not store extracted frames permanently.

## Authentication

The first version may use an internal shared token stored in environment variables.

Example request header:

```text
X-Internal-Token: <secret>
```

The token must not be committed to Git.
