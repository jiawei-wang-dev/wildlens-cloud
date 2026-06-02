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
- tag count aggregation and primary species selection
- image thumbnail generation with aspect ratio preserved
- video frame extraction at 1 frame per second

## TODO For Real Integration

- Replace `FakeStorageClient` with GCP Cloud Storage operations.
- Add checksum-based deduplication before expensive media processing.
- Upload generated thumbnails to `thumbnails/`.
- Upload video poster or extracted frames to `video-posters/`.
- Load model configuration from `MODEL_CONFIG_URI`.
- Replace the fake detector with the selected wildlife ML model.
- Replace `FakeDbClient` with the team-agreed metadata database.
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
