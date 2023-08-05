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
sqlc:
	cmd.exe /c "docker run --rm -v ${PWD}:/src -w /src kjconroy/sqlc generate"

# generate mock db for testing
mock:
	mockgen -package mockdb -destination internal/db/mock/store.go github.com/aalug/go-gin-job-search/internal/db/sqlc Store

# generate mock functions for elasticsearch based functions
mock_es:
	mockgen -package mockesearch -destination internal/esearch/mock/search.go github.com/aalug/go-gin-job-search/internal/esearch ESearchClient


# run all tests
test:
	go test -v -cover ./...

# run tests in the given path (p) and display results in the html file
test_coverage:
	go test $(p) -coverprofile=coverage.out && go tool cover -html=coverage.out

runserver:
	go run cmd/gin-job-search/main.go

# flush db and restart it
flush_db:
	docker-compose down
	docker volume ls -qf dangling=true | xargs docker volume rm
	docker-compose up -d

flush_es:
	docker-compose stop elasticsearch
	docker-compose rm -f elasticsearch
	docker-compose up -d elasticsearch

.PHONY: generate_migrations, migrate_up, migrate_down, sqlc, mock, test, test_coverage, runserver, flush_db, flush_es
