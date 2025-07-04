from datetime import datetime, timezone
import random
import uuid

from src.config import REGIONS


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
