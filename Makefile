.PHONY: run test migrate-up migrate-down docker-up docker-down

run:
	go run cmd/api/main.go

test:
	go test -v ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	export $(cat .env | xargs) && migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	export $(cat .env | xargs) && migrate -path migrations -database "$$DATABASE_URL" down

sqlc:
	sqlc generate