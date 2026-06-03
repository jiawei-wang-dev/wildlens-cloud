from __future__ import annotations

from urllib.parse import quote


DEFAULT_MEDIA_BUCKET = "pending-media-bucket"


def generate_upload_url(bucket: str, object_path: str, content_type: str) -> str:
    """Return a fake upload URL until the final storage target and signed URL flow are agreed."""
    # TODO: Replace with the selected provider's signed upload URL generation.
    encoded_path = quote(object_path, safe="/")
    encoded_content_type = quote(content_type, safe="")
    return f"fake-upload://{bucket}/{encoded_path}?content_type={encoded_content_type}"
