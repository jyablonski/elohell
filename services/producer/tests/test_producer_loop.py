from unittest.mock import patch, MagicMock

from src.main import trigger_producer_loop


def test_trigger_logs_queue_start_and_stop(mock_user, caplog):
    mock_redis = MagicMock()

    with patch("src.main.generate_fake_user", return_value=mock_user), patch(
        "src.main.write_redis_message"
    ), patch("src.main.time.sleep", side_effect=KeyboardInterrupt):

        with caplog.at_level("INFO"):
            trigger_producer_loop(redis_client=mock_redis)

        assert "Queue producer started." in caplog.text
        assert "Queue producer stopped by user." in caplog.text
