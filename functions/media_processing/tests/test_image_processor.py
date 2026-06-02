from pathlib import Path
import sys

from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from image_processor import generate_thumbnail


def test_generate_thumbnail_preserves_aspect_ratio_and_max_size(tmp_path):
    input_path = tmp_path / "wide.png"
    output_path = tmp_path / "thumb.jpg"
    Image.new("RGB", (1200, 600), color="green").save(input_path)

    result = generate_thumbnail(str(input_path), str(output_path), max_size=(300, 300))

    assert result["width"] == 300
    assert result["height"] == 150
    assert result["output_path"] == str(output_path)
    assert result["size_bytes"] > 0

    with Image.open(output_path) as thumbnail:
        assert thumbnail.format == "JPEG"
        assert thumbnail.size == (300, 150)
