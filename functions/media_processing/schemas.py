from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, TypedDict


class Detection(TypedDict):
    label: str
    count: int
    confidence: float


@dataclass
class MediaMetadata:
    file_id: str
    owner_id: Optional[str]
    original_filename: str
    file_type: str
    mime_type: Optional[str]
    checksum_sha256: Optional[str]
    size: Optional[int]
    storage_provider: str
    bucket: str
    object_path: str
    file_url: str
    thumbnail_url: Optional[str]
    thumbnail_object_path: Optional[str]
    tags: List[str]
    tag_counts: Dict[str, int]
    primary_species: Optional[str]
    model_version: Optional[str]
    status: str
    created_at: str
    updated_at: str
    duration_seconds: Optional[float] = None
    sampled_frame_rate_fps: Optional[float] = None
    sampled_frame_count: Optional[int] = None
    frame_object_paths: List[str] = field(default_factory=list)


@dataclass
class ProcessingResult:
    processed: bool
    status: str
    bucket: Optional[str] = None
    object_name: Optional[str] = None
    file_id: Optional[str] = None
    media_metadata: Optional[MediaMetadata] = None
    detections: List[Detection] = field(default_factory=list)
    message: Optional[str] = None
    details: Dict[str, Any] = field(default_factory=dict)
