from __future__ import annotations

from dataclasses import asdict
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Sequence, Union

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
    """Aggregate image detection counts by animal label."""
    tag_counts: Dict[str, int] = {}
    for detection in detections:
        label = detection["label"]
        count = int(detection.get("count", 1))
        tag_counts[label] = tag_counts.get(label, 0) + count
    return tag_counts


def aggregate_video_frame_tag_counts(frame_detections: Iterable[Iterable[Detection]]) -> Dict[str, int]:
    """Aggregate video detections using the max species count seen in any sampled frame."""
    tag_counts: Dict[str, int] = {}
    for detections in frame_detections:
        frame_tag_counts = aggregate_tag_counts(detections)
        for label, count in frame_tag_counts.items():
            tag_counts[label] = max(tag_counts.get(label, 0), count)
    return tag_counts


def choose_primary_species(tag_counts: Dict[str, int]) -> Optional[str]:
    """Choose the species with the highest count, using name order for ties."""
    if not tag_counts:
        return None
    return sorted(tag_counts.items(), key=lambda item: (-item[1], item[0]))[0][0]


def build_final_object_path(primary_species: Optional[str], checksum_sha256: str, filename: str) -> str:
    species_path = primary_species or "unknown"
    return f"media/originals/{species_path}/{checksum_sha256}/{filename}"


def build_thumbnail_object_path(checksum_sha256: str) -> str:
    return f"media/thumbnails/{checksum_sha256}.jpg"


def process_event(
    event: dict,
    storage_client: Optional[FakeStorageClient] = None,
    detector=detect_image,
    db_client: Optional[FakeDbClient] = None,
) -> dict:
    """Legacy local helper for tests; production uses Cloud Run POST /infer."""
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

    checksum_sha256 = event.get("metadata", {}).get("checksum_sha256")
    file_id = checksum_sha256 or event.get("metadata", {}).get("file_id") or Path(object_name).stem
    now = datetime.now(timezone.utc).isoformat()
    mime_type = event.get("contentType")
    file_type = _infer_file_type(object_name, mime_type)

    if file_type == "unknown":
        result = ProcessingResult(
            processed=False,
            status="failed",
            bucket=bucket,
            object_name=object_name,
            file_id=file_id,
            message="Unsupported media file type.",
        )
        return asdict(result)

    local_path = storage_client.download_object(bucket, object_name)

    # TODO: Move final orchestration to the AWS Lambda coordinator and Cloud Run /infer flow.
    # TODO: For images, generate and upload a thumbnail through coordinator-provided URLs.
    # TODO: For videos, extract 1 FPS frames and upload a poster/frames through agreed URLs.
    # TODO: Add checksum-based deduplication before expensive media processing.
    raw_detections = detector(local_path)
    sampled_frame_count = None
    if file_type == "video":
        frame_detections = _normalize_video_frame_detections(raw_detections)
        tag_counts = aggregate_video_frame_tag_counts(frame_detections)
        detections = _flatten_frame_detections(frame_detections)
        sampled_frame_count = len(frame_detections)
    else:
        detections = raw_detections
        tag_counts = aggregate_tag_counts(detections)
    primary_species = choose_primary_species(tag_counts)
    original_filename = Path(object_name).name
    final_object_path = build_final_object_path(primary_species, file_id, original_filename)
    thumbnail_object_path = build_thumbnail_object_path(file_id) if file_type == "image" else None

    metadata = MediaMetadata(
        file_id=file_id,
        owner_id=event.get("metadata", {}).get("owner_id"),
        original_filename=original_filename,
        file_type=file_type,
        mime_type=mime_type,
        checksum_sha256=checksum_sha256,
        size=_parse_size(event.get("size")),
        storage_provider="s3",
        bucket=bucket or "",
        object_path=final_object_path,
        file_url=f"s3://{bucket}/{final_object_path}",
        thumbnail_url=f"s3://{bucket}/{thumbnail_object_path}" if thumbnail_object_path else None,
        thumbnail_object_path=thumbnail_object_path,
        tags=sorted(tag_counts),
        tag_counts=tag_counts,
        primary_species=primary_species,
        model_version=MODEL_VERSION,
        status="ready",
        created_at=now,
        updated_at=now,
        sampled_frame_rate_fps=1.0 if file_type == "video" else None,
        sampled_frame_count=sampled_frame_count,
    )

    saved_metadata = db_client.save_media_metadata(metadata)

    result = ProcessingResult(
        processed=True,
        status="ready",
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
            "sampled_frame_count": sampled_frame_count,
            "final_object_path": final_object_path,
            "thumbnail_object_path": thumbnail_object_path,
        },
    )
    return asdict(result)


def _normalize_video_frame_detections(
    raw_detections: Union[Sequence[Detection], Sequence[Sequence[Detection]]],
) -> List[List[Detection]]:
    """Accept fake detector output or future per-frame detector output."""
    if not raw_detections:
        return []

    first_item = raw_detections[0]
    if isinstance(first_item, dict):
        return [list(raw_detections)]  # type: ignore[list-item]

    return [list(frame) for frame in raw_detections]  # type: ignore[union-attr]


def _flatten_frame_detections(frame_detections: Iterable[Iterable[Detection]]) -> List[Detection]:
    return [detection for detections in frame_detections for detection in detections]


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
