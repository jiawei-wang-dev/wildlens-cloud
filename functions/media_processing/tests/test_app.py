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


def test_infer_file_returns_ready_with_detected_tags(monkeypatch):
    client = TestClient(media_app.app)
    calls = {}

    def fake_detect_image(image_path):
        calls["image_path"] = image_path
        assert Path(image_path).exists()
        assert Path(image_path).read_bytes() == b"fake image bytes"
        return [{"label": "Bos_taurus", "count": 1, "confidence": 0.92}]

    monkeypatch.setattr(media_app, "detect_image", fake_detect_image)
    monkeypatch.setattr(media_app, "get_model_version", lambda: "provided-aussie-ecolense-v1")

    response = client.post(
        "/infer-file",
        files={"file": ("query.jpg", b"fake image bytes", "image/jpeg")},
    )

    assert response.status_code == 200
    body = response.json()
    assert body == {
        "tags": ["Bos_taurus"],
        "tag_counts": {"Bos_taurus": 1},
        "primary_species": "Bos_taurus",
        "model_version": "provided-aussie-ecolense-v1",
        "status": "ready",
        "error": None,
    }
    assert calls["image_path"].endswith("query.jpg")


def test_infer_file_detection_failure_returns_failed_response(monkeypatch):
    client = TestClient(media_app.app)

    def fake_detect_image(image_path):
        raise RuntimeError("detector unavailable")

    monkeypatch.setattr(media_app, "detect_image", fake_detect_image)
    monkeypatch.setattr(media_app, "get_model_version", lambda: "provided-aussie-ecolense-v1")

    response = client.post(
        "/infer-file",
        files={"file": ("query.jpg", b"fake image bytes", "image/jpeg")},
    )

    assert response.status_code == 200
    body = response.json()
    assert body["tags"] == []
    assert body["tag_counts"] == {}
    assert body["primary_species"] is None
    assert body["model_version"] == "provided-aussie-ecolense-v1"
    assert body["status"] == "failed"
    assert body["error"] == "query file inference failed: detector unavailable"


def make_infer_payload(**overrides):
    payload = {
        "file_id": VALID_CHECKSUM,
        "bucket": "fit5225-wildlife-media",
        "object_path": f"incoming/fake-user-id/{VALID_CHECKSUM}/koala.jpg",
        "filename": "koala.jpg",
        "file_type": "image",
        "mime_type": "image/jpeg",
        "checksum_sha256": VALID_CHECKSUM,
        "download_url": None,
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


def test_infer_image_with_download_url_uploads_thumbnail(monkeypatch, tmp_path):
    client = TestClient(media_app.app)
    calls = {}
    downloaded_path = tmp_path / "koala.jpg"
    downloaded_path.write_bytes(b"fake image bytes")

    def fake_download_media(download_url, target_dir, filename):
        calls["download"] = {
            "download_url": download_url,
            "target_dir": target_dir,
            "filename": filename,
        }
        return downloaded_path

    def fake_generate_thumbnail(input_path, output_path):
        calls["thumbnail"] = {
            "input_path": input_path,
            "output_path": output_path,
        }
        Path(output_path).write_bytes(b"fake thumbnail bytes")
        return {
            "width": 300,
            "height": 200,
            "output_path": output_path,
            "size_bytes": 20,
        }

    monkeypatch.setattr(media_app.media_downloader, "download_media", fake_download_media)
    monkeypatch.setattr(media_app, "generate_thumbnail", fake_generate_thumbnail)

    def fake_put(upload_url, data, headers, timeout):
        calls["upload"] = {
            "upload_url": upload_url,
            "data": data,
            "headers": headers,
            "timeout": timeout,
        }
        return type("Response", (), {"status_code": 200})()

    monkeypatch.setattr(media_app.requests, "put", fake_put)

    response = client.post(
        "/infer",
        json=make_infer_payload(download_url="https://example.test/download"),
    )

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ready"
    assert body["tags"] == ["koala"]
    assert body["tag_counts"] == {"koala": 1}
    assert body["thumbnail_object_path"] == f"media/thumbnails/{VALID_CHECKSUM}.jpg"
    assert calls["download"]["download_url"] == "https://example.test/download"
    assert calls["download"]["filename"] == "koala.jpg"
    assert calls["thumbnail"]["input_path"] == str(downloaded_path)
    assert calls["thumbnail"]["output_path"].endswith(f"{VALID_CHECKSUM}-thumbnail.jpg")
    assert calls["upload"] == {
        "upload_url": "https://example.test/thumbnail-upload",
        "data": b"fake thumbnail bytes",
        "headers": {"Content-Type": "image/jpeg"},
        "timeout": 15,
    }


def test_infer_image_thumbnail_upload_failure_returns_failed_response(monkeypatch, tmp_path):
    client = TestClient(media_app.app)
    downloaded_path = tmp_path / "koala.jpg"
    downloaded_path.write_bytes(b"fake image bytes")

    def fake_download_media(download_url, target_dir, filename):
        return downloaded_path

    def fake_generate_thumbnail(input_path, output_path):
        Path(output_path).write_bytes(b"fake thumbnail bytes")
        return {
            "width": 300,
            "height": 200,
            "output_path": output_path,
            "size_bytes": 20,
        }

    def fake_put(upload_url, data, headers, timeout):
        return type("Response", (), {"status_code": 403})()

    monkeypatch.setattr(media_app.media_downloader, "download_media", fake_download_media)
    monkeypatch.setattr(media_app, "generate_thumbnail", fake_generate_thumbnail)
    monkeypatch.setattr(media_app.requests, "put", fake_put)

    response = client.post(
        "/infer",
        json=make_infer_payload(download_url="https://example.test/download"),
    )

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "failed"
    assert body["tags"] == []
    assert body["tag_counts"] == {}
    assert body["primary_species"] is None
    assert body["thumbnail_object_path"] is None
    assert "thumbnail upload failed with HTTP status 403" in body["error"]


def test_infer_image_download_failure_returns_failed_response(monkeypatch):
    client = TestClient(media_app.app)

    def fake_download_media(download_url, target_dir, filename):
        raise RuntimeError("download failed")

    monkeypatch.setattr(media_app.media_downloader, "download_media", fake_download_media)

    response = client.post(
        "/infer",
        json=make_infer_payload(download_url="https://example.test/download"),
    )

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "failed"
    assert body["tags"] == []
    assert body["tag_counts"] == {}
    assert body["primary_species"] is None
    assert body["thumbnail_object_path"] is None
    assert "download failed" in body["error"]


def test_infer_without_download_url_keeps_placeholder_response(monkeypatch):
    client = TestClient(media_app.app)

    def fail_if_called(download_url, target_dir, filename):
        raise AssertionError("download_media should not be called without download_url")

    monkeypatch.setattr(media_app.media_downloader, "download_media", fail_if_called)

    response = client.post("/infer", json=make_infer_payload(download_url=None))

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ready"
    assert body["tags"] == ["koala"]
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


def test_infer_video_with_download_url_calls_downloader_and_extracts_frames(monkeypatch, tmp_path):
    client = TestClient(media_app.app)
    calls = {}
    downloaded_path = tmp_path / "koala.mp4"
    downloaded_path.write_bytes(b"fake video bytes")
    frame_paths = [
        str(tmp_path / "frames" / "frame_0000.jpg"),
        str(tmp_path / "frames" / "frame_0001.jpg"),
        str(tmp_path / "frames" / "frame_0002.jpg"),
    ]

    def fake_download_media(download_url, target_dir, filename):
        calls["download"] = {
            "download_url": download_url,
            "target_dir": target_dir,
            "filename": filename,
        }
        return downloaded_path

    def fake_extract_frames_1fps(video_path, output_dir):
        calls["frames"] = {
            "video_path": video_path,
            "output_dir": output_dir,
        }
        return frame_paths

    detections_by_frame = {
        frame_paths[0]: [{"label": "koala", "count": 1, "confidence": 0.9}],
        frame_paths[1]: [
            {"label": "koala", "count": 3, "confidence": 0.8},
            {"label": "kangaroo", "count": 1, "confidence": 0.7},
        ],
        frame_paths[2]: [{"label": "kangaroo", "count": 2, "confidence": 0.6}],
    }

    def fake_detect_image(image_path):
        return detections_by_frame[image_path]

    monkeypatch.setattr(media_app.media_downloader, "download_media", fake_download_media)
    monkeypatch.setattr(media_app, "extract_frames_1fps", fake_extract_frames_1fps)
    monkeypatch.setattr(media_app, "detect_image", fake_detect_image)

    def fail_if_upload_called(upload_url, data, headers, timeout):
        raise AssertionError("thumbnail upload should not be called for videos")

    monkeypatch.setattr(media_app.requests, "put", fail_if_upload_called)

    response = client.post(
        "/infer",
        json=make_infer_payload(
            filename="koala.mp4",
            object_path=f"incoming/fake-user-id/{VALID_CHECKSUM}/koala.mp4",
            file_type="video",
            mime_type="video/mp4",
            download_url="https://example.test/video-download",
        ),
    )

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ready"
    assert body["tags"] == ["kangaroo", "koala"]
    assert body["tag_counts"] == {"koala": 3, "kangaroo": 2}
    assert body["primary_species"] == "koala"
    assert body["thumbnail_object_path"] is None
    assert calls["download"]["download_url"] == "https://example.test/video-download"
    assert calls["download"]["filename"] == "koala.mp4"
    assert calls["frames"]["video_path"] == str(downloaded_path)
    assert calls["frames"]["output_dir"].endswith("frames")


def test_infer_video_download_failure_returns_failed_response(monkeypatch):
    client = TestClient(media_app.app)

    def fake_download_media(download_url, target_dir, filename):
        raise RuntimeError("video download failed")

    monkeypatch.setattr(media_app.media_downloader, "download_media", fake_download_media)

    response = client.post(
        "/infer",
        json=make_infer_payload(
            filename="koala.mp4",
            object_path=f"incoming/fake-user-id/{VALID_CHECKSUM}/koala.mp4",
            file_type="video",
            mime_type="video/mp4",
            download_url="https://example.test/video-download",
        ),
    )

    assert response.status_code == 200
    body = response.json()
    assert body["file_id"] == VALID_CHECKSUM
    assert body["status"] == "failed"
    assert body["tags"] == []
    assert body["tag_counts"] == {}
    assert body["primary_species"] is None
    assert body["thumbnail_object_path"] is None
    assert "video download failed" in body["error"]


def test_infer_video_frame_extraction_failure_returns_failed_response(monkeypatch, tmp_path):
    client = TestClient(media_app.app)
    downloaded_path = tmp_path / "koala.mp4"
    downloaded_path.write_bytes(b"fake video bytes")

    def fake_download_media(download_url, target_dir, filename):
        return downloaded_path

    def fake_extract_frames_1fps(video_path, output_dir):
        raise RuntimeError("frame extraction failed")

    monkeypatch.setattr(media_app.media_downloader, "download_media", fake_download_media)
    monkeypatch.setattr(media_app, "extract_frames_1fps", fake_extract_frames_1fps)

    response = client.post(
        "/infer",
        json=make_infer_payload(
            filename="koala.mp4",
            object_path=f"incoming/fake-user-id/{VALID_CHECKSUM}/koala.mp4",
            file_type="video",
            mime_type="video/mp4",
            download_url="https://example.test/video-download",
        ),
    )

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "failed"
    assert "frame extraction failed" in body["error"]


def test_infer_video_without_download_url_keeps_placeholder_response(monkeypatch):
    client = TestClient(media_app.app)

    def fail_if_called(download_url, target_dir, filename):
        raise AssertionError("download_media should not be called without download_url")

    monkeypatch.setattr(media_app.media_downloader, "download_media", fail_if_called)

    response = client.post(
        "/infer",
        json=make_infer_payload(
            filename="koala.mp4",
            object_path=f"incoming/fake-user-id/{VALID_CHECKSUM}/koala.mp4",
            file_type="video",
            mime_type="video/mp4",
            download_url=None,
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
