# Media Processing

This module is the first-stage local skeleton for Member B's File Handling and Model Handling work in WildLens Cloud.

## Current Scope

The code is designed to be testable locally and does not connect to real cloud services.

Implemented in this skeleton:

- GCP Cloud Storage object event routing shape in `main.py`
- `incoming/` object filtering to avoid trigger loops
- fake storage client methods for download, upload, and move
- fake database metadata save
- fake image detector
- image tag count aggregation and primary species selection
- video per-frame tag aggregation using the maximum species count across sampled frames
- image thumbnail generation with aspect ratio preserved
- video frame extraction at 1 frame per second

## Confirmed Member B to Member C Metadata Contract

Member C confirmed this contract for the future database schema. The current code keeps fake local clients and does not implement real Firestore yet.

- `file_id` is `checksum_sha256`. If a local test event omits the checksum, the skeleton falls back to legacy metadata or the object stem only so local processing can still run.
- `tag_counts` is stored as a map/object, for example `{"koala": 3, "wombat": 1}`.
- Keep both `file_url` and `thumbnail_url`.
- Also store `bucket`, `object_path`, and `thumbnail_object_path` so deletion does not depend on expiring signed URLs.
- For images, aggregate detections by summing counts for matching labels in the image.
- For videos, extract one frame per second. `tags` is the union of labels across sampled frames, and `tag_counts` stores the maximum detected count per species across sampled frames, not the sum across all frames.
- Optional video metadata fields may include `duration_seconds`, `sampled_frame_rate_fps`, `sampled_frame_count`, and `frame_object_paths`.

## TODO For Real Integration

- Replace `FakeStorageClient` with GCP Cloud Storage operations.
- Add checksum-based deduplication before expensive media processing, using `checksum_sha256`.
- Upload generated thumbnails to `thumbnails/`.
- Upload video poster or extracted frames to `video-posters/`.
- Load model configuration from `MODEL_CONFIG_URI`.
- Replace the fake detector with the selected wildlife ML model.
- Replace `FakeDbClient` with the team-agreed metadata database. Firestore is likely, but not implemented in this skeleton.
- Include Cognito owner information passed from the upload API metadata.

## Running Tests

Install dependencies for this module:

```bash
pip install -r requirements.txt
```

Run tests from this directory:

```bash
pytest
```

Video extraction uses `opencv-python`. The current tests focus on the stable local skeleton pieces; if test video generation is unreliable on a machine, keep video tests minimal until the final processing pipeline is agreed.
