run: stop
	@go run cmd/*.go

build:
	@go build -o bin/api cmd/*.go

test:
	@go test ./...

clean:
	@rm -rf bin

stop:
	@lsof -t -i:8080 | xargs kill -9 || true

watch:
	@air

migrate-up:
	@goose -env .env up

migrate-down:
	@goose -env .env down

migrate-status:
	@goose -env .env status
