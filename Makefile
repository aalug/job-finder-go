# generate migrations, $(name) - name of the migration
generate_migrations:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

# run up migrations, user details based on docker-compose.yml
migrate_up:
	migrate -path internal/db/migrations -database "postgresql://devuser:admin@localhost:5432/go_gin_job_search_db?sslmode=disable" -verbose up

# run down migrations, user details based on docker-compose.yml
migrate_down:
	migrate -path internal/db/migrations -database "postgresql://devuser:admin@localhost:5432/go_gin_job_search_db?sslmode=disable" -verbose down

# generate db related go code with sqlc
# for windows:	cmd.exe /c "docker run --rm -v ${PWD}:/src -w /src kjconroy/sqlc generate"
sqlc:
	sqlc generate

# generate mock db for testing
mock:
	mockgen -package mockdb -destination internal/db/mock/store.go github.com/aalug/go-gin-job-search/internal/db/sqlc Store

# generate mock TaskDistributor
mock_td:
	mockgen -package mockworker -destination internal/worker/mock/distributor.go github.com/aalug/go-gin-job-search/internal/worker TaskDistributor

# generate mock functions for elasticsearch based functions
mock_es:
	mockgen -package mockesearch -destination internal/esearch/mock/search.go github.com/aalug/go-gin-job-search/internal/esearch ESearchClient

# run all tests
test:
	go test -v -cover -short ./...

# run tests in the given path (p) and display results in the html file
test_coverage:
	go test $(p) -coverprofile=coverage.out && go tool cover -html=coverage.out

# run main function -> start the server.
# $(load_data): boolean - set true if you want to load test/ sample data into the postgres and elasticsearch
runserver:
	go run cmd/main.go -load_test_data=$(load_data)

# flush containers and restart it
flush:
	docker-compose down
	docker volume ls -qf dangling=true | xargs docker volume rm
	docker-compose up -d

# generate swag documentation files
swag:
	swag init -g cmd/main.go

.PHONY: generate_migrations, migrate_up, migrate_down, sqlc, mock, test, test_coverage, runserver, flush_db, flush_es, swag
