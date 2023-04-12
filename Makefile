help:
	    @echo ""
	    @echo "Makefile commands:"
	    @echo ""
	    @echo "DOCKER"
	    @echo ""
	    @echo "start           - Start the docker containers                     - ex: make start"
	    @echo "stop            - Stop the docker containers                      - ex: make stop"
	    @echo ""
	    @echo "CLIs"
	    @echo ""
	    @echo "server          - Access go container                             - ex: make server"
	    @echo "database        - Access clickhouse client                        - ex: make database"
	    @echo "test            - Run tests                                       - ex: make test"
	    @echo ""

start:
	docker-compose -f docker-compose.yml up -d

stop:
	docker-compose -f docker-compose.yml down

server:
	docker exec -it logme_server /bin/sh

database:
	docker exec -it -uroot logme_database /usr/bin/clickhouse --client -d logme

test:
	docker exec -it logme_server /bin/sh -c "go test"
