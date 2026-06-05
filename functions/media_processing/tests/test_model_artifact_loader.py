from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

import model_artifact_loader


def _write_required_artifacts(model_dir: Path) -> None:
    model_dir.mkdir(parents=True, exist_ok=True)
    for filename in model_artifact_loader.REQUIRED_ARTIFACT_FILENAMES:
        (model_dir / filename).write_bytes(b"artifact")


class FakeBlob:
    def __init__(self, object_name, downloads):
        self.object_name = object_name
        self.downloads = downloads

    def download_to_filename(self, destination):
        self.downloads.append((self.object_name, destination))
        Path(destination).write_bytes(b"downloaded artifact")


class FakeBucket:
    def __init__(self, downloads):
        self.downloads = downloads

    def blob(self, object_name):
        return FakeBlob(object_name, self.downloads)


class FakeStorageClient:
    def __init__(self, downloads):
        self.downloads = downloads

    def bucket(self, bucket_name):
        self.downloads.append(("bucket", bucket_name))
        return FakeBucket(self.downloads)


def test_ensure_model_artifacts_uses_complete_local_dir_without_gcs(tmp_path, monkeypatch):
    local_model_dir = tmp_path / "local-model"
    _write_required_artifacts(local_model_dir)
    monkeypatch.setenv("PROVIDED_MODEL_DIR", str(local_model_dir))
    monkeypatch.setenv("GCS_MODEL_BUCKET", "fit5225-wildlens-model-artifacts")
    monkeypatch.setenv("GCS_MODEL_PREFIX", "aussie-ecolense/v1")
    monkeypatch.setattr(
        model_artifact_loader,
        "_create_storage_client",
        lambda: (_ for _ in ()).throw(AssertionError("GCS should not be called")),
    )

    model_dir = model_artifact_loader.ensure_model_artifacts_available()

    assert model_dir == local_model_dir


def test_ensure_model_artifacts_downloads_from_gcs_when_local_missing(tmp_path, monkeypatch):
    downloads = []
    cache_dir = tmp_path / "cache-model"
    monkeypatch.setenv("PROVIDED_MODEL_DIR", str(tmp_path / "missing-local"))
    monkeypatch.setenv("GCS_MODEL_BUCKET", "fit5225-wildlens-model-artifacts")
    monkeypatch.setenv("GCS_MODEL_PREFIX", "aussie-ecolense/v1")
    monkeypatch.setenv("MODEL_ARTIFACT_CACHE_DIR", str(cache_dir))
    monkeypatch.setattr(
        model_artifact_loader,
        "_create_storage_client",
        lambda: FakeStorageClient(downloads),
    )

    model_dir = model_artifact_loader.ensure_model_artifacts_available()

    assert model_dir == cache_dir
    assert all((cache_dir / filename).is_file() for filename in model_artifact_loader.DOWNLOAD_ARTIFACT_FILENAMES)
    assert ("bucket", "fit5225-wildlens-model-artifacts") in downloads
    assert ("aussie-ecolense/v1/mdv5a.pt", str(cache_dir / "mdv5a.pt")) in downloads
    assert ("aussie-ecolense/v1/model.pt", str(cache_dir / "model.pt")) in downloads
    assert ("aussie-ecolense/v1/labels.txt", str(cache_dir / "labels.txt")) in downloads
    assert ("aussie-ecolense/v1/config.yaml", str(cache_dir / "config.yaml")) in downloads


def test_ensure_model_artifacts_reuses_complete_cache_without_gcs(tmp_path, monkeypatch):
    cache_dir = tmp_path / "cache-model"
    _write_required_artifacts(cache_dir)
    monkeypatch.setenv("PROVIDED_MODEL_DIR", str(tmp_path / "missing-local"))
    monkeypatch.setenv("GCS_MODEL_BUCKET", "fit5225-wildlens-model-artifacts")
    monkeypatch.setenv("GCS_MODEL_PREFIX", "aussie-ecolense/v1")
    monkeypatch.setenv("MODEL_ARTIFACT_CACHE_DIR", str(cache_dir))
    monkeypatch.setattr(
        model_artifact_loader,
        "_create_storage_client",
        lambda: (_ for _ in ()).throw(AssertionError("GCS should not be called")),
    )

    model_dir = model_artifact_loader.ensure_model_artifacts_available()

    assert model_dir == cache_dir


def test_ensure_model_artifacts_returns_none_when_gcs_download_fails(tmp_path, monkeypatch):
    monkeypatch.setenv("PROVIDED_MODEL_DIR", str(tmp_path / "missing-local"))
    monkeypatch.setenv("GCS_MODEL_BUCKET", "fit5225-wildlens-model-artifacts")
    monkeypatch.setenv("GCS_MODEL_PREFIX", "aussie-ecolense/v1")
    monkeypatch.setenv("MODEL_ARTIFACT_CACHE_DIR", str(tmp_path / "cache-model"))
    monkeypatch.setattr(
        model_artifact_loader,
        "_create_storage_client",
        lambda: (_ for _ in ()).throw(RuntimeError("GCS unavailable")),
    )

    model_dir = model_artifact_loader.ensure_model_artifacts_available()

    assert model_dir is None
