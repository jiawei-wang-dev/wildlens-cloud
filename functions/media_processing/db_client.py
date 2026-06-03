from __future__ import annotations

from typing import Dict

from models import MediaMetadata


class FakeDbClient:
    """In-memory placeholder for future metadata database writes."""

    def __init__(self) -> None:
        self.records: Dict[str, MediaMetadata] = {}

    def save_media_metadata(self, metadata: MediaMetadata) -> MediaMetadata:
        # TODO: Replace with Firestore or the team-agreed metadata database client.
        self.records[metadata.file_id] = metadata
        return metadata


_DEFAULT_DB = FakeDbClient()


def save_media_metadata(metadata: MediaMetadata) -> MediaMetadata:
    return _DEFAULT_DB.save_media_metadata(metadata)
