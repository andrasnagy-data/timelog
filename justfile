default:
    just --list

# Start PostgreSQL container
start-db:
    docker compose up -d postgres

# Start database and run the server
start: start-db
    export $(cat .env.dev | xargs) \
    && go run .

# Stop PostgreSQL container
stop-db:
    docker compose down

# Run migrations
migrate-up env="dev":
    export $(cat .env.{{env}} | xargs) \
    && migrate -path migrations -database "$DATABASE_URL" up

# Roll back migrations
migrate-down env="dev":
    export $(cat .env.{{env}} | xargs) \
    && migrate -path migrations -database "$DATABASE_URL" down 1

# Create new migration files
migrate-create name:
    migrate create -ext sql -dir migrations -seq {{name}}
