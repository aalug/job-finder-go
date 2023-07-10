# generate migrations, $(name) - name of the migration
generate_migrations:
	migrate create -ext sql -dir db/migrations -seq $(name)

# run up migrations, user details based on docker-compose.yml
migrate_up:
	migrate -path db/migrations -database "postgresql://devuser:admin@localhost:5432/go_gin_job_search_db?sslmode=disable" -verbose up

# run down migrations, user details based on docker-compose.yml
migrate_down:
	migrate -path db/migrations -database "postgresql://devuser:admin@localhost:5432/go_gin_job_search_db?sslmode=disable" -verbose down

# generate db related go code with sqlc
sqlc:
	cmd.exe /c "docker run --rm -v ${PWD}:/src -w /src kjconroy/sqlc generate"

# generate mock db for testing
mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/aalug/go-gin-job-search/db/sqlc Store

# run all tests
test:
	go test -v -cover ./...

# run tests in the given path (p) and display results in the html file
test_coverage:
	go test $(p) -coverprofile=coverage.out && go tool cover -html=coverage.out

runserver:
	go run main.go

# flush db and restart it
flush_db:
	docker-compose down
	docker volume ls -qf dangling=true | xargs docker volume rm
	docker-compose up -d

.PHONY: generate_migrations, migrate_up, migrate_down, sqlc, mock, test, test_coverage, runserver, flush_db
