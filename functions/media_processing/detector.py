from __future__ import annotations

from typing import List

from models import Detection


MODEL_VERSION = "fake-detector-v0"


def detect_image(image_path: str) -> List[Detection]:
    """Return deterministic fake detections for local skeleton testing."""
    # TODO: Load model configuration from MODEL_CONFIG_URI before real inference.
    # TODO: Replace this fake detector with the agreed wildlife ML model.
    return [
        {
            "label": "koala",
            "count": 1,
            "confidence": 0.9,
        }
    ]
