from src.config import QUEUE_NAME
from src.redis_client import write_redis_message


def test_queue_order(redis_fixture):
    redis_fixture.delete(QUEUE_NAME)
    write_redis_message(
        redis_client=redis_fixture, queue_name=QUEUE_NAME, message={"user_id": "first"}
    )
    write_redis_message(
        redis_client=redis_fixture, queue_name=QUEUE_NAME, message={"user_id": "second"}
    )

    second_out = redis_fixture.lpop(QUEUE_NAME)
    first_out = redis_fixture.lpop(QUEUE_NAME)

    assert "second" in second_out
    assert "first" in first_out
