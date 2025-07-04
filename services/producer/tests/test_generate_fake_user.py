from src.config import REGIONS
from src.user_generator import generate_fake_user


def test_generate_fake_user():
    user = generate_fake_user()
    assert len(user["user_id"]) == 36
    assert user["elo"] >= 800 and user["elo"] <= 2800
    assert user["region"] in REGIONS
    assert user["queued_at"].endswith("Z")  # ISO format with 'Z' for UTC
