from __future__ import annotations

from pathlib import Path

import requests


class MediaDownloadError(RuntimeError):
    """Raised when a coordinator-provided media URL cannot be downloaded."""


def _validate_filename(filename: str) -> str:
    if not filename or filename in {".", ".."}:
        raise ValueError("filename must be a plain file name")

    candidate = Path(filename)
    if candidate.name != filename or "/" in filename or "\\" in filename:
        raise ValueError("filename must not contain path separators or traversal")

    return filename


def download_media(download_url: str, target_dir: Path, filename: str) -> Path:
    """Download media from a temporary URL into the provided target directory."""
    safe_filename = _validate_filename(filename)
    target_dir.mkdir(parents=True, exist_ok=True)
    target_path = target_dir / safe_filename

    try:
        with requests.get(download_url, stream=True, timeout=15) as response:
            if not 200 <= response.status_code < 300:
                raise MediaDownloadError(
                    f"media download failed with HTTP status {response.status_code}"
                )

            with target_path.open("wb") as output_file:
                for chunk in response.iter_content(chunk_size=1024 * 1024):
                    if chunk:
                        output_file.write(chunk)
    except MediaDownloadError:
        raise
    except requests.RequestException as exc:
        raise MediaDownloadError(f"media download request failed: {exc}") from exc
    except OSError as exc:
        raise MediaDownloadError(f"media download file write failed: {exc}") from exc

    return target_path
