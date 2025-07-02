from datetime import datetime, timezone
import json
import logging
import random
import uuid
import time

import redis

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)

redis_client = redis.Redis(host="redis", port=6379, decode_responses=True)

REGIONS = ["NA", "EU", "AUS", "ASIA"]


def generate_fake_user():
    # generate random elo rating
    # using a normal distribution centered around 1500 with a standard deviation of 300
    # with min and max between 800 and 2800
    elo = int(random.gauss(1500, 300))
    elo = max(800, min(elo, 2800))

    return {
        "user_id": str(uuid.uuid4()),
        "elo": elo,
        "region": random.choice(REGIONS),
        "queued_at": datetime.now(timezone.utc).isoformat() + "Z",
    }


def produce():
    logging.info("Queue producer started.")
    try:
        while True:
            user = generate_fake_user()
            redis_client.lpush("match_queue", json.dumps(user))
            logging.info(f"Queued user: {user}")
            time.sleep(random.uniform(0.5, 2.0))
    except KeyboardInterrupt:
        logging.info("Queue producer stopped by user.")


if __name__ == "__main__":
    produce()
