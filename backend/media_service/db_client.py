from __future__ import annotations

from dataclasses import dataclass
from typing import Dict, Optional


@dataclass
class ExistingMediaRecord:
    file_id: str
    checksum_sha256: str
    file_url: str
    thumbnail_url: Optional[str] = None


class FakeMediaDbClient:
    """In-memory placeholder for future Firestore-backed checksum lookups."""

    def __init__(self, records: Optional[Dict[str, ExistingMediaRecord]] = None) -> None:
        self.records: Dict[str, ExistingMediaRecord] = records or {}

    def find_by_checksum(self, checksum_sha256: str) -> Optional[ExistingMediaRecord]:
        # TODO: Replace with Firestore lookup keyed by checksum_sha256/file_id.
        return self.records.get(checksum_sha256)


_DEFAULT_DB = FakeMediaDbClient()


def find_by_checksum(checksum_sha256: str) -> Optional[ExistingMediaRecord]:
    return _DEFAULT_DB.find_by_checksum(checksum_sha256)
