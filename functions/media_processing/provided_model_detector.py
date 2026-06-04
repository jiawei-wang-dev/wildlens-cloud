from __future__ import annotations

from dataclasses import dataclass
import os
from pathlib import Path
import tempfile
from typing import Any, List

try:
    from .models import Detection
except ImportError:
    from models import Detection


DEFAULT_MODEL_DIR = Path(__file__).resolve().parent / "model_artifacts" / "AussieEcoLense"
MODEL_DIR_ENV = "PROVIDED_MODEL_DIR"
MEGADETECTOR_FILENAME = "mdv5a.pt"
CLASSIFIER_FILENAME = "model.pt"
LABELS_FILENAME = "labels.txt"
ANIMAL_CATEGORY = "1"
LOWER_CONFIDENCE_THRESHOLD = 0.05
SNIP_SIZE = 600


@dataclass
class _ProvidedModelBundle:
    megadetector: Any
    classifier: Any
    transform: Any
    labels: list[str]
    device: str
    torch: Any
    load_image: Any


_MODEL_CACHE: _ProvidedModelBundle | None = None
_MODEL_CACHE_DIR: Path | None = None
_MODEL_DIR_OVERRIDE: Path | None = None


def get_provided_model_dir() -> Path:
    """Return the configured local directory for the provided model package."""
    if _MODEL_DIR_OVERRIDE is not None:
        return _MODEL_DIR_OVERRIDE

    return Path(os.environ.get(MODEL_DIR_ENV, DEFAULT_MODEL_DIR))


def configure_model_dir(model_dir: Path) -> None:
    """Point the provided detector at a resolved model directory."""
    global _MODEL_CACHE, _MODEL_CACHE_DIR, _MODEL_DIR_OVERRIDE

    resolved_model_dir = Path(model_dir)
    if _MODEL_CACHE_DIR is not None and _MODEL_CACHE_DIR != resolved_model_dir:
        _MODEL_CACHE = None
        _MODEL_CACHE_DIR = None

    _MODEL_DIR_OVERRIDE = resolved_model_dir


def provided_model_files_available(model_dir: Path | None = None) -> bool:
    """Return whether the required local model artifacts are present."""
    resolved_model_dir = model_dir or get_provided_model_dir()
    required_files = (
        resolved_model_dir / MEGADETECTOR_FILENAME,
        resolved_model_dir / CLASSIFIER_FILENAME,
        resolved_model_dir / LABELS_FILENAME,
    )
    return all(path.is_file() for path in required_files)


def detect_image_with_provided_model(image_path: Path) -> List[Detection]:
    """Detect animals with MegaDetector and classify each crop with the provided model."""
    bundle = _load_model_bundle()
    image_path = Path(image_path)
    md_result = _run_megadetector(bundle, image_path)

    with tempfile.TemporaryDirectory() as temporary_dir:
        crop_dir = Path(temporary_dir) / "cropped_images"
        crop_dir.mkdir(parents=True, exist_ok=True)
        crop_paths = _crop_animal_detections(md_result, image_path, crop_dir)

        detections: list[Detection] = []
        for crop_path in crop_paths:
            label, confidence = _classify_crop(bundle, crop_path)
            detections.append(
                {
                    "label": label,
                    "count": 1,
                    "confidence": confidence,
                }
            )

    return detections


def _load_model_bundle() -> _ProvidedModelBundle:
    global _MODEL_CACHE, _MODEL_CACHE_DIR

    if _MODEL_CACHE is not None:
        return _MODEL_CACHE

    model_dir = get_provided_model_dir()
    if not provided_model_files_available(model_dir):
        raise FileNotFoundError(f"provided model artifacts are missing in {model_dir}")

    from megadetector.detection.run_detector import load_detector
    from megadetector.visualization.visualization_utils import load_image
    import torch
    import torchvision.transforms as transforms

    device = _select_device(torch)
    megadetector = load_detector(str(model_dir / MEGADETECTOR_FILENAME), force_cpu=device == "cpu")
    classifier = _load_classifier(torch, model_dir / CLASSIFIER_FILENAME, device)
    transform = transforms.Compose(
        [
            transforms.Resize((480, 480)),
            transforms.ToTensor(),
        ]
    )

    _MODEL_CACHE = _ProvidedModelBundle(
        megadetector=megadetector,
        classifier=classifier,
        transform=transform,
        labels=_load_labels(model_dir / LABELS_FILENAME),
        device=device,
        torch=torch,
        load_image=load_image,
    )
    _MODEL_CACHE_DIR = model_dir
    return _MODEL_CACHE


def _select_device(torch: Any) -> str:
    if torch.cuda.is_available():
        return "cuda"

    mps_backend = getattr(torch.backends, "mps", None)
    if mps_backend is not None and mps_backend.is_available():
        return "mps"

    return "cpu"


def _load_classifier(torch: Any, model_path: Path, device: str) -> Any:
    try:
        model = torch.load(str(model_path), map_location=device, weights_only=False)
    except TypeError:
        model = torch.load(str(model_path), map_location=device)

    model.eval()
    model.to(device)
    return model


def _run_megadetector(bundle: _ProvidedModelBundle, image_path: Path) -> dict[str, Any]:
    image = bundle.load_image(str(image_path))
    result = bundle.megadetector.generate_detections_one_image(
        image,
        str(image_path),
        detection_threshold=LOWER_CONFIDENCE_THRESHOLD,
        verbose=False,
    )

    failure = result.get("failure")
    if failure:
        raise RuntimeError(f"MegaDetector failed for {image_path}: {failure}")

    return result


def _crop_animal_detections(md_result: dict[str, Any], image_path: Path, crop_dir: Path) -> list[Path]:
    from PIL import Image

    crop_paths: list[Path] = []

    with Image.open(image_path) as source_image:
        image = source_image.convert("RGB")
        width, height = image.size

        for detection in md_result.get("detections", []):
            if str(detection.get("category")) != ANIMAL_CATEGORY:
                continue

            confidence = float(detection.get("conf", 0.0))
            if confidence < LOWER_CONFIDENCE_THRESHOLD:
                continue

            crop_box = _normalized_bbox_to_pixels(detection["bbox"], width, height)
            if crop_box is None:
                continue

            crop = image.crop(crop_box)
            resized_crop = crop.resize((SNIP_SIZE, SNIP_SIZE), Image.BILINEAR)
            crop_path = crop_dir / f"{image_path.stem}-{len(crop_paths)}{image_path.suffix}"
            resized_crop.save(crop_path)
            crop_paths.append(crop_path)

    return crop_paths


def _normalized_bbox_to_pixels(
    bbox: list[float],
    image_width: int,
    image_height: int,
) -> tuple[int, int, int, int] | None:
    x, y, width, height = bbox
    left = max(0, min(image_width, int(x * image_width)))
    top = max(0, min(image_height, int(y * image_height)))
    right = max(0, min(image_width, int((x + width) * image_width)))
    bottom = max(0, min(image_height, int((y + height) * image_height)))

    if right <= left or bottom <= top:
        return None

    return left, top, right, bottom


def _classify_crop(bundle: _ProvidedModelBundle, crop_path: Path) -> tuple[str, float]:
    from PIL import Image

    with Image.open(crop_path) as image:
        tensor = bundle.transform(image.convert("RGB"))

    tensor = tensor.unsqueeze(0)
    tensor = tensor.permute(0, 2, 3, 1)
    tensor = tensor.to(bundle.device)

    with bundle.torch.no_grad():
        logits = bundle.classifier(tensor)
        probabilities = bundle.torch.softmax(logits, dim=1)[0].detach().cpu()

    best_index = int(bundle.torch.argmax(probabilities).item())
    label = bundle.labels[best_index] if best_index < len(bundle.labels) else f"class_{best_index}"
    confidence = float(probabilities[best_index].item())
    return label, confidence


def _load_labels(labels_path: Path) -> list[str]:
    labels: list[str] = []

    for raw_line in labels_path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line:
            continue

        fields = line.split(";")
        if len(fields) >= 6:
            genus = fields[4].strip()
            species = fields[5].strip()
            if genus and species:
                labels.append(f"{genus.capitalize()}_{species.lower()}")
            elif genus:
                labels.append(genus.capitalize())
            else:
                labels.append(line)
        else:
            labels.append(line)

    if not labels:
        raise ValueError(f"no labels found in {labels_path}")

    return labels
