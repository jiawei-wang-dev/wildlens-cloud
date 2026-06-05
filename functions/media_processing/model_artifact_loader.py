from __future__ import annotations

import logging
import os
from pathlib import Path
from typing import Any


DEFAULT_LOCAL_MODEL_DIR = Path(__file__).resolve().parent / "model_artifacts" / "AussieEcoLense"
DEFAULT_CACHE_DIR = Path("/tmp/aussie-ecolense/v1")
PROVIDED_MODEL_DIR_ENV = "PROVIDED_MODEL_DIR"
GCS_MODEL_BUCKET_ENV = "GCS_MODEL_BUCKET"
GCS_MODEL_PREFIX_ENV = "GCS_MODEL_PREFIX"
MODEL_ARTIFACT_CACHE_DIR_ENV = "MODEL_ARTIFACT_CACHE_DIR"
REQUIRED_ARTIFACT_FILENAMES = ("mdv5a.pt", "model.pt", "labels.txt")
DOWNLOAD_ARTIFACT_FILENAMES = REQUIRED_ARTIFACT_FILENAMES + ("config.yaml",)

logger = logging.getLogger(__name__)


def ensure_model_artifacts_available() -> Path | None:
    """Return a local model directory, downloading GCS artifacts to cache when needed."""
    local_model_dir = Path(os.environ.get(PROVIDED_MODEL_DIR_ENV, DEFAULT_LOCAL_MODEL_DIR))
    if _required_artifacts_exist(local_model_dir):
        return local_model_dir

    bucket_name = os.environ.get(GCS_MODEL_BUCKET_ENV)
    prefix = os.environ.get(GCS_MODEL_PREFIX_ENV)
    if not bucket_name or not prefix:
        return None

    cache_dir = Path(os.environ.get(MODEL_ARTIFACT_CACHE_DIR_ENV, DEFAULT_CACHE_DIR))
    if _required_artifacts_exist(cache_dir):
        return cache_dir

    try:
        _download_gcs_artifacts(bucket_name, prefix, cache_dir)
    except Exception:
        logger.warning("Failed to download model artifacts from GCS.", exc_info=True)
        return None

    if _required_artifacts_exist(cache_dir):
        return cache_dir

    return None


def _required_artifacts_exist(model_dir: Path) -> bool:
    return model_dir.is_dir() and all((model_dir / filename).is_file() for filename in REQUIRED_ARTIFACT_FILENAMES)


def _download_gcs_artifacts(bucket_name: str, prefix: str, cache_dir: Path) -> None:
    cache_dir.mkdir(parents=True, exist_ok=True)
    clean_prefix = prefix.strip("/")
    bucket = _create_storage_client().bucket(bucket_name)

    for filename in DOWNLOAD_ARTIFACT_FILENAMES:
        object_name = f"{clean_prefix}/{filename}" if clean_prefix else filename
        destination = cache_dir / filename
        bucket.blob(object_name).download_to_filename(str(destination))


def _create_storage_client() -> Any:
    from google.cloud import storage

    return storage.Client()
