from __future__ import annotations

from pathlib import Path
from typing import List


def extract_frames_1fps(video_path: str, output_dir: str) -> List[str]:
    """Extract one frame per second from a video file."""
    # TODO: In production, decide whether frames are stored as processing artifacts,
    # thumbnails, or both, and persist their URLs in database metadata as needed.
    try:
        import cv2
    except ImportError as exc:
        raise RuntimeError("opencv-python is required for video frame extraction") from exc

    output = Path(output_dir)
    output.mkdir(parents=True, exist_ok=True)

    capture = cv2.VideoCapture(video_path)
    if not capture.isOpened():
        raise ValueError(f"Could not open video file: {video_path}")

    fps = capture.get(cv2.CAP_PROP_FPS)
    frame_interval = max(int(round(fps)), 1)
    saved_frames: List[str] = []
    frame_index = 0
    second_index = 0

    try:
        while True:
            success, frame = capture.read()
            if not success:
                break

            if frame_index % frame_interval == 0:
                frame_path = output / f"frame_{second_index:04d}.jpg"
                cv2.imwrite(str(frame_path), frame)
                saved_frames.append(str(frame_path))
                second_index += 1

            frame_index += 1
    finally:
        capture.release()

    return saved_frames
