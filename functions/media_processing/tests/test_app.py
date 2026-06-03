from pathlib import Path
import sys

from fastapi.testclient import TestClient

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

import app as media_app


VALID_CHECKSUM = "c" * 64


def test_health_returns_ok():
    client = TestClient(media_app.app)

    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def make_infer_payload(**overrides):
    payload = {
        "file_id": VALID_CHECKSUM,
        "bucket": "fit5225-wildlife-media",
        "object_path": f"incoming/fake-user-id/{VALID_CHECKSUM}/koala.jpg",
        "filename": "koala.jpg",
        "file_type": "image",
        "mime_type": "image/jpeg",
        "checksum_sha256": VALID_CHECKSUM,
        "download_url": "https://example.test/download",
        "thumbnail_upload_url": "https://example.test/thumbnail-upload",
    }
    payload.update(overrides)
    return payload


def test_infer_accepts_valid_image_request_and_returns_ready_result():
    client = TestClient(media_app.app)

    response = client.post("/infer", json=make_infer_payload())

    assert response.status_code == 200
    body = response.json()
    assert body["file_id"] == VALID_CHECKSUM
    assert body["tags"] == ["koala"]
    assert body["tag_counts"] == {"koala": 1}
    assert body["primary_species"] == "koala"
    assert body["status"] == "ready"
    assert body["thumbnail_object_path"] == f"media/thumbnails/{VALID_CHECKSUM}.jpg"


def test_infer_returns_ready_for_supported_video_placeholder():
    client = TestClient(media_app.app)

    response = client.post(
        "/infer",
        json=make_infer_payload(
            filename="koala.mp4",
            object_path=f"incoming/fake-user-id/{VALID_CHECKSUM}/koala.mp4",
            file_type="video",
            mime_type="video/mp4",
        ),
    )

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ready"
    assert body["tag_counts"] == {"koala": 1}
    assert body["thumbnail_object_path"] is None


def test_infer_does_not_write_dynamodb_directly():
    client = TestClient(media_app.app)

    response = client.post("/infer", json=make_infer_payload())

    assert response.status_code == 200
    assert "db_client" not in media_app.__dict__
    assert "dynamodb" not in media_app.__dict__


def test_infer_unsupported_file_type_fails_validation():
    client = TestClient(media_app.app)

    response = client.post("/infer", json=make_infer_payload(file_type="audio"))

    assert response.status_code == 422


def test_infer_rejects_file_id_that_does_not_match_checksum():
    client = TestClient(media_app.app)

    response = client.post("/infer", json=make_infer_payload(file_id="d" * 64))

    assert response.status_code == 422
