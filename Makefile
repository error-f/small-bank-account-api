.PHONY: run clean-run test

# Set the Docker Compose files
COMPOSE_FILES = -f docker-compose.yml -f docker-compose.test.yml

# Run the application
run:
	docker-compose -f docker-compose.yml up --build

clean-run:
	rm -rf pgdata
	docker-compose -f docker-compose.yml up --build

# Run tests with the test database
test:
	docker-compose -f docker-compose.test.yml up -d
	go test
	docker-compose -f docker-compose.test.yml down
