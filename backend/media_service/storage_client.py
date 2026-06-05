from __future__ import annotations

import os
from urllib.parse import quote


DEFAULT_MEDIA_BUCKET_FALLBACK = "pending-media-bucket"
DEFAULT_STORAGE_PROVIDER_FALLBACK = "s3"
DEFAULT_AWS_REGION_FALLBACK = "us-east-1"


def get_media_bucket() -> str:
    return os.getenv("MEDIA_BUCKET", DEFAULT_MEDIA_BUCKET_FALLBACK)


def get_media_storage_provider() -> str:
    return os.getenv("MEDIA_STORAGE_PROVIDER", DEFAULT_STORAGE_PROVIDER_FALLBACK)


def get_aws_region() -> str:
    return os.getenv("AWS_REGION", DEFAULT_AWS_REGION_FALLBACK)


def generate_upload_url(bucket: str, object_path: str, content_type: str) -> str:
    """Return a fake upload URL until AWS S3 presigned uploads are integrated."""
    # TODO: Generate an AWS S3 presigned upload URL using AWS_REGION and MEDIA_BUCKET.
    encoded_path = quote(object_path, safe="/")
    encoded_content_type = quote(content_type, safe="")
    return f"fake-upload://{bucket}/{encoded_path}?content_type={encoded_content_type}"
