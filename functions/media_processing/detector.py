from __future__ import annotations

import logging
import os
from pathlib import Path
from typing import List

try:
    from . import provided_model_detector
    from .models import Detection
except ImportError:
    import provided_model_detector
    from models import Detection


FAKE_MODEL_VERSION = "fake-detector-v0"
PROVIDED_MODEL_VERSION = "provided-aussie-ecolense-v1"
MODEL_VERSION = FAKE_MODEL_VERSION
USE_PROVIDED_MODEL_ENV = "USE_PROVIDED_MODEL"

logger = logging.getLogger(__name__)


def detect_image(image_path: str) -> List[Detection]:
    """Run the configured image detector, falling back to deterministic fake output."""
    if _should_use_provided_model():
        try:
            detections = provided_model_detector.detect_image_with_provided_model(Path(image_path))
            _set_model_version(PROVIDED_MODEL_VERSION)
            return detections
        except Exception:
            logger.warning("Provided model detector failed; falling back to fake detector.", exc_info=True)

    _set_model_version(FAKE_MODEL_VERSION)
    return _detect_image_with_fake_detector()


def get_model_version() -> str:
    """Return the detector version used by the latest successful detector route."""
    return MODEL_VERSION


def _should_use_provided_model() -> bool:
    if not _environment_flag_enabled(os.environ.get(USE_PROVIDED_MODEL_ENV)):
        return False

    return provided_model_detector.provided_model_files_available()


def _environment_flag_enabled(value: str | None) -> bool:
    return value is not None and value.strip().lower() in {"1", "true", "yes", "on"}


def _set_model_version(model_version: str) -> None:
    global MODEL_VERSION
    MODEL_VERSION = model_version


def _detect_image_with_fake_detector() -> List[Detection]:
    """Return deterministic fake detections for local skeleton testing."""
    return [
        {
            "label": "koala",
            "count": 1,
            "confidence": 0.9,
        }
    ]
