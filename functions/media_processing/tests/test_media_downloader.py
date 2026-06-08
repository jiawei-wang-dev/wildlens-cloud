from pathlib import Path
import sys

import pytest
import requests

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from media_downloader import MediaDownloadError, download_media


class FakeResponse:
    def __init__(self, status_code=200, chunks=None):
        self.status_code = status_code
        self._chunks = chunks or [b"wild", b"lens"]

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, traceback):
        return False

    def iter_content(self, chunk_size):
        return iter(self._chunks)


def test_download_media_writes_file_to_target_directory(tmp_path, monkeypatch):
    def fake_get(url, stream, timeout):
        assert url == "https://example.test/media.jpg"
        assert stream is True
        assert timeout == 15
        return FakeResponse(chunks=[b"koala", b"-image"])

    monkeypatch.setattr("media_downloader.requests.get", fake_get)

    downloaded_path = download_media(
        "https://example.test/media.jpg",
        tmp_path,
        "koala.jpg",
    )

    assert downloaded_path == tmp_path / "koala.jpg"
    assert downloaded_path.read_bytes() == b"koala-image"


def test_download_media_raises_clear_error_for_http_error(tmp_path, monkeypatch):
    monkeypatch.setattr(
        "media_downloader.requests.get",
        lambda url, stream, timeout: FakeResponse(status_code=403),
    )

    with pytest.raises(MediaDownloadError, match="HTTP status 403"):
        download_media("https://example.test/forbidden.jpg", tmp_path, "koala.jpg")


@pytest.mark.parametrize(
    "filename",
    [
        "../koala.jpg",
        "nested/koala.jpg",
        r"nested\koala.jpg",
        "..",
        "",
    ],
)
def test_download_media_rejects_unsafe_filename(tmp_path, filename):
    with pytest.raises(ValueError, match="filename"):
        download_media("https://example.test/media.jpg", tmp_path, filename)


def test_download_media_wraps_timeout_or_network_error(tmp_path, monkeypatch):
    def fake_get(url, stream, timeout):
        raise requests.Timeout("request timed out")

    monkeypatch.setattr("media_downloader.requests.get", fake_get)

    with pytest.raises(MediaDownloadError, match="request failed"):
        download_media("https://example.test/slow.jpg", tmp_path, "koala.jpg")
