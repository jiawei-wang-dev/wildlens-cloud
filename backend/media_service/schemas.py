from __future__ import annotations

from typing import Literal, Optional

from pydantic import BaseModel, Field


class UploadUrlRequest(BaseModel):
    filename: str = Field(..., min_length=1)
    content_type: str = Field(..., min_length=1)
    size: int = Field(..., gt=0)
    checksum_sha256: str = Field(..., min_length=1)
    file_type: Literal["image", "video"]


class UploadUrlResponse(BaseModel):
    duplicate: bool
    file_id: str
    message: str
    existing_file_url: Optional[str] = None
    existing_thumbnail_url: Optional[str] = None
    upload_url: Optional[str] = None
    bucket: Optional[str] = None
    object_path: Optional[str] = None
