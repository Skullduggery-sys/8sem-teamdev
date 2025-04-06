include .env

ifeq ($(POSTGRES_SETUP_PROD),)
	POSTGRES_SETUP_PROD := user=${DB_USERNAME_RW} password=${DB_PASSWORD_RW} dbname=postgres host=localhost port=5436 sslmode=disable
endif
ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=${DB_USERNAME_RW} password=${DB_PASSWORD_RW} dbname=postgres_test host=localhost port=5436 sslmode=disable
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
# setup-replication:
# # Create replication user on master
# 	docker exec -it go-postgres-replication_postgres-master_1 psql -U user -d mydb -c "CREATE ROLE replication WITH LOGIN REPLICATION PASSWORD 'password';"

# # Restart master to apply config changes
# 	docker-compose restart postgres-master

# # Initialize slave with base backup
# 	docker exec -it go-postgres-replication_postgres-slave_1 pg_basebackup -h postgres-master -U replication -D /var/lib/postgresql/data -P --wal-method=stream

# # Restart slave to apply config changes
# 	docker-compose restart postgres-slave

app:
	go run cmd/main.go 1
bal1:
	go run cmd/main.go 2
bal2:
	go run cmd/main.go 3

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
# check connect: docker exec -it media-organizer-slave psql -h postgres-master -U ro_user -d postgres

swagger: $(SWAGGER_SRC)
	swag init --parseDependency --parseInternal -g ./cmd/main.go -o ./swagger --exclude ./internal/api/v1

.PHONY: fake-tests lint intergation-tests swagger
