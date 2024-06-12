.PHONY: help list start stop server test clickhouse psql valkey tail

# Default target
all: start

## Display this help message
help:
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} \
		/^[a-zA-Z0-9_-]+:.*##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

list: ## List the docker containers
	docker compose -f docker-compose.yml ps

start: ## Start the docker containers
	docker compose -f docker-compose.yml up -d

restart: ## Restart and rebuild the docker containers
	docker compose -f docker-compose.yml up --remove-orphans --build --force-recreate -d

stop: ## Stop the docker containers
	docker compose -f docker-compose.yml down

server: ## Access the server container
	docker exec -it logme_server /bin/sh

queue: ## Access the queue container
	docker exec -it logme_queue /bin/sh

test: ## Run tests
	docker exec -it logme_server /bin/sh -c "go test"

clickhouse: ## Access the ClickHouse CLI
	docker compose -f docker-compose.yml exec logs /usr/bin/clickhouse --client -d logs

psql: ## Access the PostgreSQL CLI
	docker compose -f docker-compose.yml exec database psql -U admin -d main

valkey: ## Access the Valkey CLI
	docker compose -f docker-compose.yml exec cache valkey-cli

tail: ## Tail the app logs
	$(eval ID := $(shell docker ps --filter "name=logme_server" -q))
	docker logs -f ${ID}

tail_queue_stdout: ## Tail the queue stdout logs
	docker compose -f docker-compose.yml exec queue tail -f /var/log/queue.out.log

tail_queue_stderr: ## Tail the queue stderr logs
	docker compose -f docker-compose.yml exec queue tail -f /var/log/queue.err.log
