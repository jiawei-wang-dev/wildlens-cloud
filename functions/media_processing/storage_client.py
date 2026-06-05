from __future__ import annotations

from pathlib import Path
from tempfile import gettempdir
from typing import Dict, Optional


class FakeStorageClient:
    """Local placeholder used by legacy object-event tests."""

    def __init__(self) -> None:
        self.uploaded_objects: Dict[str, str] = {}
        self.moved_objects: Dict[str, str] = {}

    def download_object(self, bucket: str, object_name: str, destination_path: Optional[str] = None) -> str:
        """Return a local placeholder path without contacting object storage."""
        # TODO: The Cloud Run service should use coordinator-provided download URLs.
        if destination_path:
            return destination_path
        return str(Path(gettempdir()) / Path(object_name).name)

    def upload_object(self, bucket: str, source_path: str, object_name: str) -> str:
        """Record an upload operation and return a fake public URL."""
        # TODO: The Cloud Run service should use coordinator-provided upload URLs.
        self.uploaded_objects[object_name] = source_path
        return f"fake://{bucket}/{object_name}"

    def move_object(self, bucket: str, source_object_name: str, destination_object_name: str) -> str:
        """Record a move operation and return the fake destination URL."""
        # TODO: Final object moves belong in the AWS Lambda coordinator workflow.
        self.moved_objects[source_object_name] = destination_object_name
        return f"fake://{bucket}/{destination_object_name}"


def download_object(bucket: str, object_name: str, destination_path: Optional[str] = None) -> str:
    return FakeStorageClient().download_object(bucket, object_name, destination_path)


def upload_object(bucket: str, source_path: str, object_name: str) -> str:
    return FakeStorageClient().upload_object(bucket, source_path, object_name)


def move_object(bucket: str, source_object_name: str, destination_object_name: str) -> str:
    return FakeStorageClient().move_object(bucket, source_object_name, destination_object_name)
