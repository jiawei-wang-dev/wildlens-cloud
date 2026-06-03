from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from db_client import FakeDbClient
from main import (
    aggregate_tag_counts,
    aggregate_video_frame_tag_counts,
    build_final_object_path,
    build_thumbnail_object_path,
    choose_primary_species,
    process_event,
    should_process_object,
)
from storage_client import FakeStorageClient


VALID_CHECKSUM = "b" * 64


def test_incoming_prefix_is_processed():
    assert should_process_object("incoming/user-1/koala.jpg") is True


def test_output_prefixes_are_ignored():
    assert should_process_object("media/user-1/koala.jpg") is False
    assert should_process_object("thumbnails/user-1/koala.jpg") is False
    assert should_process_object("video-posters/user-1/clip.jpg") is False
    assert should_process_object("models/wildlife-model.json") is False
    assert should_process_object("failed/user-1/koala.jpg") is False


def test_aggregate_tag_counts_sums_matching_labels():
    detections = [
        {"label": "koala", "count": 1, "confidence": 0.9},
        {"label": "kangaroo", "count": 2, "confidence": 0.8},
        {"label": "koala", "count": 3, "confidence": 0.7},
    ]

    assert aggregate_tag_counts(detections) == {"koala": 4, "kangaroo": 2}


def test_aggregate_video_frame_tag_counts_uses_max_per_species_across_frames():
    frame_detections = [
        [{"label": "koala", "count": 1, "confidence": 0.9}],
        [{"label": "koala", "count": 3, "confidence": 0.8}],
        [
            {"label": "koala", "count": 2, "confidence": 0.7},
            {"label": "kangaroo", "count": 1, "confidence": 0.85},
        ],
    ]

    assert aggregate_video_frame_tag_counts(frame_detections) == {"koala": 3, "kangaroo": 1}


def test_choose_primary_species_uses_highest_count():
    tag_counts = {"koala": 1, "kangaroo": 4, "wombat": 2}

    assert choose_primary_species(tag_counts) == "kangaroo"


def test_path_builders_create_stage_3_storage_paths():
    assert (
        build_final_object_path("koala", VALID_CHECKSUM, "koala.jpg")
        == f"media/originals/koala/{VALID_CHECKSUM}/koala.jpg"
    )
    assert build_thumbnail_object_path(VALID_CHECKSUM) == f"media/thumbnails/{VALID_CHECKSUM}.jpg"


def test_process_event_uses_checksum_sha256_as_file_id_and_ready_status():
    db_client = FakeDbClient()
    event = {
        "bucket": "wildlens-test",
        "name": "incoming/user-1/koala.jpg",
        "contentType": "image/jpeg",
        "metadata": {
            "file_id": "legacy-file-id",
            "checksum_sha256": VALID_CHECKSUM,
        },
    }

    result = process_event(
        event,
        storage_client=FakeStorageClient(),
        detector=lambda _path: [{"label": "koala", "count": 1, "confidence": 0.9}],
        db_client=db_client,
    )

    assert result["file_id"] == VALID_CHECKSUM
    assert result["status"] == "ready"
    assert result["media_metadata"]["file_id"] == VALID_CHECKSUM
    assert result["media_metadata"]["status"] == "ready"
    assert VALID_CHECKSUM in db_client.records


def test_image_metadata_has_thumbnail_object_path_and_final_object_path():
    event = {
        "bucket": "wildlens-test",
        "name": "incoming/user-1/koala.jpg",
        "contentType": "image/jpeg",
        "metadata": {"checksum_sha256": VALID_CHECKSUM},
    }

    result = process_event(
        event,
        storage_client=FakeStorageClient(),
        detector=lambda _path: [{"label": "koala", "count": 1, "confidence": 0.9}],
        db_client=FakeDbClient(),
    )

    metadata = result["media_metadata"]
    expected_object_path = f"media/originals/koala/{VALID_CHECKSUM}/koala.jpg"
    expected_thumbnail_path = f"media/thumbnails/{VALID_CHECKSUM}.jpg"
    assert metadata["object_path"] == expected_object_path
    assert metadata["object_path"] != event["name"]
    assert metadata["thumbnail_object_path"] == expected_thumbnail_path
    assert metadata["file_url"] == f"s3://wildlens-test/{expected_object_path}"
    assert metadata["thumbnail_url"] == f"s3://wildlens-test/{expected_thumbnail_path}"


def test_unknown_file_type_does_not_call_detector_and_fails_safely():
    event = {
        "bucket": "wildlens-test",
        "name": "incoming/user-1/readme.txt",
        "contentType": "text/plain",
        "metadata": {"checksum_sha256": VALID_CHECKSUM},
    }

    def detector(_path):
        raise AssertionError("detector should not be called for unknown file types")

    result = process_event(
        event,
        storage_client=FakeStorageClient(),
        detector=detector,
        db_client=FakeDbClient(),
    )

    assert result["processed"] is False
    assert result["status"] == "failed"
    assert result["file_id"] == VALID_CHECKSUM
    assert result["media_metadata"] is None
