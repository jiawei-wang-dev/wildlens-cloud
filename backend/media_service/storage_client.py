from __future__ import annotations

import os
from urllib.parse import quote


DEFAULT_MEDIA_BUCKET_FALLBACK = "pending-media-bucket"
DEFAULT_STORAGE_PROVIDER_FALLBACK = "gcp"


def get_media_bucket() -> str:
    return os.getenv("MEDIA_BUCKET", DEFAULT_MEDIA_BUCKET_FALLBACK)


def get_media_storage_provider() -> str:
    return os.getenv("MEDIA_STORAGE_PROVIDER", DEFAULT_STORAGE_PROVIDER_FALLBACK)


def generate_upload_url(bucket: str, object_path: str, content_type: str) -> str:
    """Return a fake upload URL until the final storage target and signed URL flow are agreed."""
    # TODO: Replace with the selected provider's signed upload URL generation.
    encoded_path = quote(object_path, safe="/")
    encoded_content_type = quote(content_type, safe="")
    return f"fake-upload://{bucket}/{encoded_path}?content_type={encoded_content_type}"
