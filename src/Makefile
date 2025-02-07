include .env

ifeq ($(POSTGRES_SETUP_PROD),)
	POSTGRES_SETUP_PROD := user=${DB_USERNAME} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT} sslmode=disable
endif
# ifeq ($(POSTGRES_SETUP_TEST),)
# 	POSTGRES_SETUP_TEST := user=${DB_USERNAME} password=${DB_PASSWORD} dbname=postgres_test host=localhost port=5436 sslmode=disable
# endif

SWAGGER_SRC := $(wildcard ./internal/controller/*.go)

MIGRATION_FOLDER=./migrations

migration-up:
	@goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_PROD)" up
migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_PROD)" down

docker-build:
	docker compose -f docker-compose.yml build
docker-up:
	docker compose -f docker-compose.yml up postgres --build --detach
docker-down:
	docker compose -f docker-compose.yml down
docker-killdb:
	docker compose -f docker-compose.yml down --volumes

docker-app-up:
	docker compose -f docker-compose.yml up app --build --detach

# ci-unit-tests-coverage:
# 	docker compose -f docker-compose.yml run app make unit-tests-coverage
# ci-integration-tests:
# 	docker compose -f docker-compose.yml run app make integration-tests

build:
	go build -o main cmd/main.go

app:
	go run cmd/main.go

unit-tests:
	go test ./... -count=1 -shuffle=on -json | go-test-report
unit-tests-coverage:
	go test -covermode=atomic -coverprofile=coverage/coverage.out.tmp -count=1 -shuffle=on ./...
	cat coverage/coverage.out.tmp | grep -v "_mock.go" > coverage/coverage.out
	go tool cover -html=coverage/coverage.out -o coverage/coverage_report.html
integration-tests:
	go test -tags=integration ./internal/tests -v -count=1 -shuffle=on
e2e-setup:
	curl -X POST -H "Content-Type: application/json" \
	-d '{"name":"alexey vasilyev","login":"alivasilyev","email":"convex.hull.trick@mail.ru","role":"user","password":"12345"}' \
	http://${APP_HOST}:9000/sign-up
	curl -X POST "http://${APP_HOST}:9000/verify-2fa?code=228228&email=convex.hull.trick@mail.ru"
e2e-tests:
	go test -tags=e2e ./e2e -v -count=1
lint:
	golangci-lint run --config ./golangci.yaml

jaeger-up:
	docker run -d -p 6831:6831/udp -p 16686:16686 jaegertracing/all-in-one:latest

jaeger-down:
	docker stop

swagger: $(SWAGGER_SRC)
	swag init --parseDependency --parseInternal -g ./cmd/main.go -o ./swagger

.PHONY: unit-tests unit-tests-coverage lint integration-tests swagger
