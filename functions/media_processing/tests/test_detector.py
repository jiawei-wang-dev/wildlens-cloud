from pathlib import Path
import sys

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

import detector
from detector import detect_image


@pytest.fixture(autouse=True)
def reset_detector_environment(monkeypatch):
    monkeypatch.delenv("USE_PROVIDED_MODEL", raising=False)
    monkeypatch.delenv("PROVIDED_MODEL_DIR", raising=False)
    monkeypatch.setattr(detector, "MODEL_VERSION", detector.FAKE_MODEL_VERSION)


def test_detect_image_fake_detector_returns_expected_structure(tmp_path, monkeypatch):
    image_path = tmp_path / "koala.jpg"
    image_path.write_bytes(b"fake image bytes")
    monkeypatch.setattr(
        detector.provided_model_detector,
        "detect_image_with_provided_model",
        lambda _path: pytest.fail("provided model should not be called"),
    )

    detections = detect_image(str(image_path))

    assert isinstance(detections, list)
    assert detections
    assert detections[0]["label"] == "koala"
    assert detections[0]["count"] == 1
    assert detections[0]["confidence"] == 0.9
    assert detector.get_model_version() == detector.FAKE_MODEL_VERSION


def test_detect_image_falls_back_when_provided_model_files_are_missing(tmp_path, monkeypatch):
    image_path = tmp_path / "koala.jpg"
    image_path.write_bytes(b"fake image bytes")
    missing_model_dir = tmp_path / "missing-model-dir"
    monkeypatch.setenv("USE_PROVIDED_MODEL", "true")
    monkeypatch.setenv("PROVIDED_MODEL_DIR", str(missing_model_dir))
    monkeypatch.setattr(
        detector.provided_model_detector,
        "detect_image_with_provided_model",
        lambda _path: pytest.fail("provided model should not be called"),
    )

    detections = detect_image(str(image_path))

    assert detections == [{"label": "koala", "count": 1, "confidence": 0.9}]
    assert detector.get_model_version() == detector.FAKE_MODEL_VERSION


def test_detect_image_uses_mocked_provided_model_when_enabled(tmp_path, monkeypatch):
    image_path = tmp_path / "brush_turkey.jpg"
    image_path.write_bytes(b"fake image bytes")
    expected_detections = [{"label": "Alectura_lathami", "count": 1, "confidence": 1.0}]

    monkeypatch.setenv("USE_PROVIDED_MODEL", "true")
    monkeypatch.setattr(
        detector.provided_model_detector,
        "provided_model_files_available",
        lambda: True,
    )
    monkeypatch.setattr(
        detector.provided_model_detector,
        "detect_image_with_provided_model",
        lambda path: expected_detections if path == image_path else [],
    )

    detections = detect_image(str(image_path))

    assert detections == expected_detections
    assert detector.get_model_version() == detector.PROVIDED_MODEL_VERSION


def test_detect_image_falls_back_when_provided_model_load_fails(tmp_path, monkeypatch):
    image_path = tmp_path / "koala.jpg"
    image_path.write_bytes(b"fake image bytes")

    monkeypatch.setenv("USE_PROVIDED_MODEL", "true")
    monkeypatch.setattr(
        detector.provided_model_detector,
        "provided_model_files_available",
        lambda: True,
    )
    monkeypatch.setattr(
        detector.provided_model_detector,
        "detect_image_with_provided_model",
        lambda _path: (_ for _ in ()).throw(RuntimeError("model load failed")),
    )

    detections = detect_image(str(image_path))

    assert detections == [{"label": "koala", "count": 1, "confidence": 0.9}]
    assert detector.get_model_version() == detector.FAKE_MODEL_VERSION
