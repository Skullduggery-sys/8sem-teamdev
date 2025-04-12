include .env

ifeq ($(POSTGRES_SETUP_PROD),)
	POSTGRES_SETUP_PROD := user=${DB_USERNAME} password=${DB_PASSWORD} dbname=postgres host=localhost port=${DB_PORT} sslmode=disable
endif
ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=${DB_USERNAME} password=${DB_PASSWORD} dbname=postgres_test host=localhost port=${DB_PORT} sslmode=disable
endif

SWAGGER_SRC := $(wildcard ./internal/controller/v1/*.go)

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

fake-tests:
	go test ./internal/service -cover -count=1
intergation-tests:
	go test -tags=integration ./internal/tests/ -v -count=1
lint:
	golangci-lint run --config ./golangci.yaml

nginx-up:
	sudo nginx -c ~/Folders/Stuff/university/7sem/web/7sem-web/src/nginx/nginx.conf
nginx-down:
	sudo nginx -c ~/Folders/Stuff/university/7sem/web/7sem-web/src/nginx/nginx.conf -s stop
nginx-logs:
	gedit /var/log/nginx/access.log
nginx-errors:
	gedit /var/log/nginx/error.log
nginx-test:
	k6 run ./nginx/test/script.js

# misc:
# sudo netstat -lpn |grep :9080
# sudo truncate --size 0 /var/log/nginx/access.log

swagger: $(SWAGGER_SRC)
	swag init --parseDependency --parseInternal -g ./cmd/main.go -o ./swagger --exclude ./internal/api/v1

.PHONY: fake-tests lint intergation-tests swagger
