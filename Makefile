# generate migrations, $(name) - name of the migration
generate_migrations:
	migrate create -ext sql -dir db/migrations -seq $(name)

# run up migrations, user details based on docker-compose.yml
migrate_up:
	migrate -path db/migrations -database "postgresql://devuser:admin@localhost:5432/go_gin_job_search_db?sslmode=disable" -verbose up

# run down migrations, user details based on docker-compose.yml
migrate_down:
	migrate -path db/migrations -database "postgresql://devuser:admin@localhost:5432/go_gin_job_search_db?sslmode=disable" -verbose down
