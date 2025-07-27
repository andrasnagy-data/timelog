default:
    just --list

# Start PostgreSQL container
start-db:
    docker compose up -d postgres

# Start database and run the server
start: start-db swagger-gen
    export $(cat env/.env.dev | xargs) \
    && go run ./cmd/server

# Stop PostgreSQL container
stop-db:
    docker compose down

# Run migrations
migrate-up env="dev":
    export $(cat env/.env.{{env}} | xargs) \
    && migrate -path migrations -database "$DATABASE_URL" up

# Roll back migrations
migrate-down env="dev":
    export $(cat env/.env.{{env}} | xargs) \
    && migrate -path migrations -database "$DATABASE_URL" down 1

# Create new migration files
migrate-create name:
    migrate create -ext sql -dir migrations -seq {{name}}

# Generate Swagger documentation from code annotations
swagger-gen:
    swag init -g cmd/server/main.go -o api/
