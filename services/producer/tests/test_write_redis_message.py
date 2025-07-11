from src.redis_client import write_redis_message
from src.config import QUEUE_NAME


def test_write_message(redis_fixture):
    data = "123"
    message = {"user_id": data}
    write_redis_message(
        redis_client=redis_fixture, queue_name=QUEUE_NAME, message=message
    )
    result = redis_fixture.lpop(QUEUE_NAME)
    assert data in result
