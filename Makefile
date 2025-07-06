.PHONY: up
up:
	@docker compose -f docker/docker-compose-local.yml up -d

.PHONY: down
down:
	@docker compose -f docker/docker-compose-local.yml down

.PHONY: build
build:
	@docker compose -f docker/docker-compose-local.yml build

.PHONY: run-producer-tests
run-producer-tests:
	@docker compose -f docker/docker-compose-producer-test.yml down
	@docker compose -f docker/docker-compose-producer-test.yml up --exit-code-from producer_service_test_runner

.PHONY: run-matchmaking-tests
run-matchmaking-tests:
	@docker compose -f docker/docker-compose-matchmaking-test.yml up --abort-on-container-exit