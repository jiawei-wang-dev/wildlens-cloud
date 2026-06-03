from __future__ import annotations

import re
from typing import Literal, Optional

from pydantic import BaseModel, Field, field_validator


SHA256_HEX_PATTERN = re.compile(r"^[0-9a-fA-F]{64}$")


class UploadUrlRequest(BaseModel):
    filename: str = Field(..., min_length=1)
    content_type: str = Field(..., min_length=1)
    size: int = Field(..., gt=0)
    checksum_sha256: str = Field(..., pattern=SHA256_HEX_PATTERN.pattern)
    file_type: Literal["image", "video"]

    @field_validator("filename")
    @classmethod
    def filename_must_be_safe(cls, value: str) -> str:
        filename = value.strip()
        if not filename:
            raise ValueError("filename cannot be blank")
        if filename in {".", ".."}:
            raise ValueError("filename cannot be a path segment")
        if "/" in filename or "\\" in filename:
            raise ValueError("filename cannot include path separators")
        return filename


class UploadUrlResponse(BaseModel):
    duplicate: bool
    file_id: str
    message: str
    existing_file_url: Optional[str] = None
    existing_thumbnail_url: Optional[str] = None
    upload_url: Optional[str] = None
    bucket: Optional[str] = None
    object_path: Optional[str] = None
