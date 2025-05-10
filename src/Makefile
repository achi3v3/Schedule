.PHONY: build run test clean docker-build docker-up docker-down

# Build the application
build:
	go build -o bin/schedule_bot main.go

# Run the application
run: build
	./bin/schedule_bot

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker-compose build

# Start Docker containers
docker-up:
	docker-compose up -d

# Stop Docker containers
docker-down:
	docker-compose down

# Initialize database
init-db:
	docker-compose exec postgres psql -U postgres -d schedule_bot -f /docker-entrypoint-initdb.d/init.sql 