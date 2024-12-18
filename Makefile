build:
	@go build -o bin/bot ./cmd/bot/main.go

run: build
	@./bin/bot --config=./config/local.yaml

test:
	@go test -v ./...

migrate:
	@go run ./cmd/migrate/main.go --config=./config/local.yaml --migrations-path=./migrations
