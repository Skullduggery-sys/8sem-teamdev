include .env

ifeq ($(POSTGRES_SETUP_PROD),)
	POSTGRES_SETUP_PROD := user=${DB_USERNAME} password=${DB_PASSWORD} dbname=postgres host=localhost port=5436 sslmode=disable
endif
ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=${DB_USERNAME} password=${DB_PASSWORD} dbname=postgres_test host=localhost port=5436 sslmode=disable
endif

SWAGGER_SRC := $(wildcard ./internal/controller/*.go)

MIGRATION_FOLDER=./migrations

migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_PROD)" up
migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_PROD)" down

docker-up:
	docker-compose -f docker-compose.yml up --build -d
docker-down:
	docker-compose -f docker-compose.yml down
docker-killdb:
	docker-compose -f docker-compose.yml down --volumes

app:
	go run cmd/main.go

unit-tests:
	go test ./... -count=1 -shuffle=on -json | go-test-report
unit-tests-coverage:
	go test -covermode=atomic -coverprofile=coverage/coverage.out.tmp -count=1 -shuffle=on ./...
	cat coverage/coverage.out.tmp | grep -v "_mock.go" > coverage/coverage.out
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
intergation-tests:
	go test -tags=integration ./internal/tests/ -v -count=1
lint:
	golangci-lint run --config ./golangci.yaml

swagger: $(SWAGGER_SRC)
	swag init --parseDependency --parseInternal -g ./cmd/main.go -o ./swagger

.PHONY: fake-tests lint intergation-tests swagger
