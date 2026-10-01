-include .env
export

MIGRATE_IMAGE=migrate/migrate

.PHONY: run build test lint migrate-up migrate-down docker-up docker-down

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -race -cover

lint:
	go vet ./...

migrate-up:
	docker run --rm -v $(PWD)/migrations:/migrations --network host $(MIGRATE_IMAGE) \
		-path=/migrations -database "$(DATABASE_URL)" up

migrate-down:
	docker run --rm -v $(PWD)/migrations:/migrations --network host $(MIGRATE_IMAGE) \
		-path=/migrations -database "$(DATABASE_URL)" down 1

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
