.PHONY: run clean-run test

# Set the Docker Compose files
COMPOSE_FILES = -f docker-compose.yml -f docker-compose.test.yml

# Run the application
run:
	docker-compose -f docker-compose.yml up --build

clean-run:
	rm -rf pgdata
	make run

# Run tests with the test database
test:
	docker-compose -f docker-compose.test.yml up -d
	./wait-for-it.sh localhost:5432 --timeout=30 -- echo "Postgres is ready!"
	go test || exit 0
	docker-compose -f docker-compose.test.yml down
