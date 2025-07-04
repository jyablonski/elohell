import logging
import random
import time

import redis

from src.config import QUEUE_NAME
from src.logging_config import configure_logging
from src.redis_client import create_redis_client, write_redis_message
from src.user_generator import generate_fake_user


def trigger_producer_loop(redis_client: redis.Redis):
    logging.info("Queue producer started.")
    try:
        while True:
            user = generate_fake_user()
            write_redis_message(
                redis_client=redis_client, queue_name=QUEUE_NAME, message=user
            )
            logging.info(f"Queued user: {user['user_id']}")
            time.sleep(random.uniform(0.5, 2.0))
    except KeyboardInterrupt:
        logging.info("Queue producer stopped by user.")


if __name__ == "__main__":
    configure_logging()
    redis_client = create_redis_client()
    trigger_producer_loop(redis_client=redis_client)
