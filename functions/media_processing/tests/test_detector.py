from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from detector import detect_image


def test_detect_image_fake_detector_returns_expected_structure(tmp_path):
    image_path = tmp_path / "koala.jpg"
    image_path.write_bytes(b"fake image bytes")

    detections = detect_image(str(image_path))

    assert isinstance(detections, list)
    assert detections
    assert detections[0]["label"] == "koala"
    assert detections[0]["count"] == 1
    assert detections[0]["confidence"] == 0.9
