.PHONY: build run test tidy fmt vet migrate-up lint

BIN := bin/nobel

build:
	go build -o $(BIN) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

# Applies the SQL migrations against $DATABASE_URL using psql.
migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is not set" && exit 1)
	psql "$(DATABASE_URL)" -f migrations/0001_init.sql
