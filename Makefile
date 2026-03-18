.PHONY: build run test lint dev-up dev-down migrate check-rls

build:
	go build -o fireline ./cmd/fireline

run: build
	./fireline

test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

dev-up:
	docker compose up -d

dev-down:
	docker compose down

migrate:
	atlas migrate hash --dir file://migrations
	atlas migrate apply --dir file://migrations --url "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"

check-rls:
	./scripts/check_rls.sh
