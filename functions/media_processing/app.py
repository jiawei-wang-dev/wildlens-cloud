from __future__ import annotations

from pathlib import Path
import tempfile
from typing import Dict, Literal, Optional

from fastapi import FastAPI
from pydantic import BaseModel, model_validator

try:
    from . import media_downloader
    from .detector import MODEL_VERSION, detect_image
    from .image_processor import generate_thumbnail
    from .main import (
        aggregate_tag_counts,
        aggregate_video_frame_tag_counts,
        build_thumbnail_object_path,
        choose_primary_species,
        _normalize_video_frame_detections,
    )
    from .video_processor import extract_frames_1fps
except ImportError:
    import media_downloader
    from detector import MODEL_VERSION, detect_image
    from image_processor import generate_thumbnail
    from main import (
        aggregate_tag_counts,
        aggregate_video_frame_tag_counts,
        build_thumbnail_object_path,
        choose_primary_species,
        _normalize_video_frame_detections,
    )
    from video_processor import extract_frames_1fps


class InferenceRequest(BaseModel):
    file_id: str
    bucket: str
    object_path: str
    filename: str
    file_type: Literal["image", "video"]
    mime_type: Optional[str] = None
    checksum_sha256: str
    download_url: Optional[str] = None
    thumbnail_upload_url: Optional[str] = None

    @model_validator(mode="after")
    def file_id_must_match_checksum(self) -> "InferenceRequest":
        if self.file_id != self.checksum_sha256:
            raise ValueError("file_id must match checksum_sha256")
        return self


class InferenceResponse(BaseModel):
    file_id: str
    tags: list[str]
    tag_counts: Dict[str, int]
    primary_species: Optional[str]
    model_version: Optional[str]
    status: str
    thumbnail_object_path: Optional[str]
    error: Optional[str] = None


app = FastAPI(title="WildLens Media Inference Service", version="0.1.0")


@app.get("/health")
def health_check() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/infer", response_model=InferenceResponse)
def infer_media(request: InferenceRequest) -> InferenceResponse:
    """Run download-based media work or return placeholder inference results."""
    thumbnail_object_path = None

    if request.file_type == "image" and request.download_url:
        try:
            with tempfile.TemporaryDirectory() as temporary_dir:
                work_dir = Path(temporary_dir)
                local_media_path = media_downloader.download_media(
                    request.download_url,
                    work_dir,
                    request.filename,
                )
                local_thumbnail_path = work_dir / f"{request.checksum_sha256}-thumbnail.jpg"
                generate_thumbnail(str(local_media_path), str(local_thumbnail_path))
                raw_detections = detect_image(str(local_media_path))
                tag_counts = aggregate_tag_counts(raw_detections)
                thumbnail_object_path = build_thumbnail_object_path(request.checksum_sha256)
        except Exception as exc:
            return InferenceResponse(
                file_id=request.file_id,
                tags=[],
                tag_counts={},
                primary_species=None,
                model_version=MODEL_VERSION,
                status="failed",
                thumbnail_object_path=None,
                error=f"image processing failed: {exc}",
            )

        primary_species = choose_primary_species(tag_counts)

        return InferenceResponse(
            file_id=request.file_id,
            tags=sorted(tag_counts),
            tag_counts=tag_counts,
            primary_species=primary_species,
            model_version=MODEL_VERSION,
            status="ready",
            thumbnail_object_path=thumbnail_object_path,
        )

    if request.file_type == "video" and request.download_url:
        try:
            with tempfile.TemporaryDirectory() as temporary_dir:
                work_dir = Path(temporary_dir)
                local_media_path = media_downloader.download_media(
                    request.download_url,
                    work_dir,
                    request.filename,
                )
                frame_paths = extract_frames_1fps(
                    str(local_media_path),
                    str(work_dir / "frames"),
                )
                frame_detections = [detect_image(frame_path) for frame_path in frame_paths]
                tag_counts = aggregate_video_frame_tag_counts(frame_detections)
        except Exception as exc:
            return InferenceResponse(
                file_id=request.file_id,
                tags=[],
                tag_counts={},
                primary_species=None,
                model_version=MODEL_VERSION,
                status="failed",
                thumbnail_object_path=None,
                error=f"video processing failed: {exc}",
            )

        primary_species = choose_primary_species(tag_counts)

        return InferenceResponse(
            file_id=request.file_id,
            tags=sorted(tag_counts),
            tag_counts=tag_counts,
            primary_species=primary_species,
            model_version=MODEL_VERSION,
            status="ready",
            thumbnail_object_path=None,
        )

    placeholder_path = request.download_url or f"s3://{request.bucket}/{request.object_path}"
    raw_detections = detect_image(placeholder_path)

    if request.file_type == "video":
        frame_detections = _normalize_video_frame_detections(raw_detections)
        tag_counts = aggregate_video_frame_tag_counts(frame_detections)
        thumbnail_object_path = None
    else:
        tag_counts = aggregate_tag_counts(raw_detections)
        thumbnail_object_path = build_thumbnail_object_path(request.checksum_sha256)

    primary_species = choose_primary_species(tag_counts)

    return InferenceResponse(
        file_id=request.file_id,
        tags=sorted(tag_counts),
        tag_counts=tag_counts,
        primary_species=primary_species,
        model_version=MODEL_VERSION,
        status="ready",
        thumbnail_object_path=thumbnail_object_path,
    )
