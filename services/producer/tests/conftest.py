import logging
import os

import pytest

from src.redis_client import create_redis_client
from src.logging_config import configure_logging


@pytest.fixture
def redis_fixture():
    if os.getenv("ENV") == "docker":
        host = "redis"
    else:
        host = "localhost"

    return create_redis_client(host=host, port=6379)


@pytest.fixture(autouse=True, scope="session")
def setup_logging():
    """
    Configure logging once for the entire test session.
    Uses the same config as the main app, but can be adjusted if needed.
    """
    configure_logging()
    logging.getLogger("redis").setLevel(logging.WARNING)  # quiet noisy libs
    logging.info("Logging configured for test session")


@pytest.fixture
def mock_user():
    return {
        "user_id": "test-user-id",
        "elo": 1500,
        "region": "NA",
        "queued_at": "2025-01-01T00:00:00Z",
    }
