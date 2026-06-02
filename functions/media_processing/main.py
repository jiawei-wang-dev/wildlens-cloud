from __future__ import annotations

from dataclasses import asdict
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, Iterable, Optional

from db_client import FakeDbClient
from detector import MODEL_VERSION, detect_image
from models import Detection, MediaMetadata, ProcessingResult
from storage_client import FakeStorageClient


PROCESS_PREFIX = "incoming/"
IGNORED_PREFIXES = (
    "media/",
    "thumbnails/",
    "video-posters/",
    "models/",
    "failed/",
)


def should_process_object(object_name: str) -> bool:
    """Return True only for new incoming objects that should trigger processing."""
    if not object_name or object_name.endswith("/"):
        return False
    if object_name.startswith(IGNORED_PREFIXES):
        return False
    return object_name.startswith(PROCESS_PREFIX)


def aggregate_tag_counts(detections: Iterable[Detection]) -> Dict[str, int]:
    """Aggregate detection counts by animal label."""
    tag_counts: Dict[str, int] = {}
    for detection in detections:
        label = detection["label"]
        count = int(detection.get("count", 1))
        tag_counts[label] = tag_counts.get(label, 0) + count
    return tag_counts


def choose_primary_species(tag_counts: Dict[str, int]) -> Optional[str]:
    """Choose the species with the highest count, using name order for ties."""
    if not tag_counts:
        return None
    return sorted(tag_counts.items(), key=lambda item: (-item[1], item[0]))[0][0]


def process_event(
    event: dict,
    storage_client: Optional[FakeStorageClient] = None,
    detector=detect_image,
    db_client: Optional[FakeDbClient] = None,
) -> dict:
    """Local skeleton for a GCP Cloud Storage object event handler."""
    storage_client = storage_client or FakeStorageClient()
    db_client = db_client or FakeDbClient()

    bucket = event.get("bucket")
    object_name = event.get("name")

    if not should_process_object(object_name or ""):
        result = ProcessingResult(
            processed=False,
            status="skipped",
            bucket=bucket,
            object_name=object_name,
            message="Object is outside the incoming/ processing prefix or is an ignored output prefix.",
        )
        return asdict(result)

    local_path = storage_client.download_object(bucket, object_name)
    file_id = event.get("metadata", {}).get("file_id") or Path(object_name).stem
    now = datetime.now(timezone.utc).isoformat()
    mime_type = event.get("contentType")
    file_type = _infer_file_type(object_name, mime_type)

    # TODO: For images, generate and upload a thumbnail to thumbnails/.
    # TODO: For videos, extract 1 FPS frames and upload a poster/frames to video-posters/.
    # TODO: Add checksum-based deduplication before expensive media processing.
    detections = detector(local_path)
    tag_counts = aggregate_tag_counts(detections)
    primary_species = choose_primary_species(tag_counts)

    metadata = MediaMetadata(
        file_id=file_id,
        owner_id=event.get("metadata", {}).get("owner_id"),
        original_filename=Path(object_name).name,
        file_type=file_type,
        mime_type=mime_type,
        checksum_sha256=event.get("metadata", {}).get("checksum_sha256"),
        size=_parse_size(event.get("size")),
        storage_provider="gcp",
        bucket=bucket or "",
        object_path=object_name,
        file_url=f"fake://{bucket}/{object_name}",
        thumbnail_url=None,
        tags=sorted(tag_counts),
        tag_counts=tag_counts,
        primary_species=primary_species,
        model_version=MODEL_VERSION,
        status="complete",
        created_at=now,
        updated_at=now,
    )

    saved_metadata = db_client.save_media_metadata(metadata)

    result = ProcessingResult(
        processed=True,
        status="complete",
        bucket=bucket,
        object_name=object_name,
        file_id=file_id,
        media_metadata=saved_metadata,
        detections=detections,
        message="Processed with local fake clients.",
        details={
            "local_path": local_path,
            "tag_counts": tag_counts,
            "primary_species": primary_species,
        },
    )
    return asdict(result)


def _infer_file_type(object_name: str, mime_type: Optional[str]) -> str:
    if mime_type:
        if mime_type.startswith("image/"):
            return "image"
        if mime_type.startswith("video/"):
            return "video"

    suffix = Path(object_name).suffix.lower()
    if suffix in {".jpg", ".jpeg", ".png", ".webp", ".gif"}:
        return "image"
    if suffix in {".mp4", ".mov", ".avi", ".mkv"}:
        return "video"
    return "unknown"


def _parse_size(size: object) -> Optional[int]:
    if size is None:
        return None
    try:
        return int(size)
    except (TypeError, ValueError):
        return None
