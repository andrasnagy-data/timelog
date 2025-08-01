# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o main ./cmd/server

# Runtime stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Create the app user
RUN addgroup -S app && adduser -S -G app app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy templates directory
COPY --from=builder /app/templates ./templates

# chown all the files to the app user
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port (Heroku will override this with PORT env var)
EXPOSE 8080

# Run the binary
CMD ["./main"]