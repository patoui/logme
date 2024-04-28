help:
	    @echo ""
	    @echo "Available commands:"
	    @echo ""
	    @echo "DOCKER"
	    @echo ""
	    @echo "list            - List the docker containers       - ex: make list"
	    @echo "start           - Start the docker containers      - ex: make start"
	    @echo "stop            - Stop the docker containers       - ex: make stop"
	    @echo ""
	    @echo "CLIs"
	    @echo ""
	    @echo "server          - Access go container              - ex: make server"
	    @echo "test            - Run tests                        - ex: make test"
	    @echo "clickhouse      - Access the clickhouse CLI        - ex: make clickhouse"
	    @echo "psql            - Access the PostgreSQL CLI        - ex: make psql"
	    @echo "valkey          - Access the Valkey CLI            - ex: make valkey"
	    @echo ""
	    @echo "HELPERS"
	    @echo ""
	    @echo "tail            - Tail the app logs                - ex: make tail"
	    @echo ""

list:
	docker compose -f docker-compose.yml ps

start:
	docker compose -f docker-compose.yml up -d

stop:
	docker compose -f docker-compose.yml down

server:
	docker exec -it logme_server /bin/sh

test:
	docker exec -it logme_server /bin/sh -c "go test"

clickhouse:
	docker compose -f docker-compose.yml exec logs /usr/bin/clickhouse --client -d logs

psql:
	docker compose -f docker-compose.yml exec database psql -U admin -d main

valkey:
	docker compose -f docker-compose.yml exec cache valkey-cli

tail:
	$(eval ID := $(shell docker ps --filter "name=logme_server" -q))
	docker logs -f ${ID}
