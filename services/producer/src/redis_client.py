import json
import logging

import redis

logger = logging.getLogger(__name__)


def create_redis_client(host: str = "redis", port: int = 6379) -> redis.Redis:
    """Create a Redis client."""
    return redis.Redis(host=host, port=port, decode_responses=True)


def write_redis_message(
    redis_client: redis.Redis, queue_name: str, message: dict[str, str]
):
    """Write a message to a Redis queue."""
    redis_client.lpush(queue_name, json.dumps(message))
    logger.info(f"Message written to queue {queue_name}: {message}")
