from __future__ import annotations

from pathlib import Path
from typing import Dict, Tuple

from PIL import Image


def generate_thumbnail(input_path: str, output_path: str, max_size: Tuple[int, int] = (300, 300)) -> Dict[str, object]:
    """Generate a JPEG thumbnail while preserving aspect ratio."""
    output = Path(output_path)
    output.parent.mkdir(parents=True, exist_ok=True)

    with Image.open(input_path) as image:
        image = image.convert("RGB")
        image.thumbnail(max_size, Image.Resampling.LANCZOS)
        image.save(output, format="JPEG", quality=85, optimize=True)
        width, height = image.size

    return {
        "width": width,
        "height": height,
        "output_path": str(output),
        "size_bytes": output.stat().st_size,
    }
