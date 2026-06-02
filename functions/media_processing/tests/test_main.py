from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from main import aggregate_tag_counts, choose_primary_species, should_process_object


def test_incoming_prefix_is_processed():
    assert should_process_object("incoming/user-1/koala.jpg") is True


def test_output_prefixes_are_ignored():
    assert should_process_object("media/user-1/koala.jpg") is False
    assert should_process_object("thumbnails/user-1/koala.jpg") is False
    assert should_process_object("video-posters/user-1/clip.jpg") is False
    assert should_process_object("models/wildlife-model.json") is False
    assert should_process_object("failed/user-1/koala.jpg") is False


def test_aggregate_tag_counts_sums_matching_labels():
    detections = [
        {"label": "koala", "count": 1, "confidence": 0.9},
        {"label": "kangaroo", "count": 2, "confidence": 0.8},
        {"label": "koala", "count": 3, "confidence": 0.7},
    ]

    assert aggregate_tag_counts(detections) == {"koala": 4, "kangaroo": 2}


def test_choose_primary_species_uses_highest_count():
    tag_counts = {"koala": 1, "kangaroo": 4, "wombat": 2}

    assert choose_primary_species(tag_counts) == "kangaroo"
