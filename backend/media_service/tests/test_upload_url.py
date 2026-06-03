from fastapi.testclient import TestClient

from backend.media_service.db_client import ExistingMediaRecord, FakeMediaDbClient
from backend.media_service.main import app, get_db_client


AUTH_HEADERS = {"Authorization": "Bearer placeholder-token"}


def make_request_payload(**overrides):
    payload = {
        "filename": "koala.jpg",
        "content_type": "image/jpeg",
        "size": 12345,
        "checksum_sha256": "abc123checksum",
        "file_type": "image",
    }
    payload.update(overrides)
    return payload


def make_client(db_client=None):
    app.dependency_overrides.clear()
    if db_client is not None:
        app.dependency_overrides[get_db_client] = lambda: db_client
    return TestClient(app)


def test_valid_request_returns_duplicate_false():
    client = make_client(FakeMediaDbClient())

    response = client.post("/media/upload-url", json=make_request_payload(), headers=AUTH_HEADERS)

    assert response.status_code == 200
    body = response.json()
    assert body["duplicate"] is False
    assert body["upload_url"]
    assert body["bucket"] == "pending-media-bucket"


def test_missing_checksum_fails_validation():
    client = make_client(FakeMediaDbClient())
    payload = make_request_payload()
    payload.pop("checksum_sha256")

    response = client.post("/media/upload-url", json=payload, headers=AUTH_HEADERS)

    assert response.status_code == 422


def test_invalid_file_type_fails_validation():
    client = make_client(FakeMediaDbClient())

    response = client.post(
        "/media/upload-url",
        json=make_request_payload(file_type="audio"),
        headers=AUTH_HEADERS,
    )

    assert response.status_code == 422


def test_duplicate_checksum_returns_duplicate_true():
    db_client = FakeMediaDbClient(
        {
            "abc123checksum": ExistingMediaRecord(
                file_id="abc123checksum",
                checksum_sha256="abc123checksum",
                file_url="fake://media/abc123checksum",
                thumbnail_url="fake://thumbnails/abc123checksum.jpg",
            )
        }
    )
    client = make_client(db_client)

    response = client.post("/media/upload-url", json=make_request_payload(), headers=AUTH_HEADERS)

    assert response.status_code == 200
    body = response.json()
    assert body["duplicate"] is True
    assert body["file_id"] == "abc123checksum"
    assert body["existing_file_url"] == "fake://media/abc123checksum"
    assert body["existing_thumbnail_url"] == "fake://thumbnails/abc123checksum.jpg"
    assert body["message"]


def test_object_path_includes_owner_id_and_checksum():
    client = make_client(FakeMediaDbClient())

    response = client.post("/media/upload-url", json=make_request_payload(), headers=AUTH_HEADERS)

    assert response.status_code == 200
    body = response.json()
    assert body["object_path"] == "incoming/fake-user-id/abc123checksum/koala.jpg"


def test_file_id_equals_checksum_sha256():
    client = make_client(FakeMediaDbClient())

    response = client.post("/media/upload-url", json=make_request_payload(), headers=AUTH_HEADERS)

    assert response.status_code == 200
    assert response.json()["file_id"] == "abc123checksum"
