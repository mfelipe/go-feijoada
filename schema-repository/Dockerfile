FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY schema-repository/go.mod schema-repository/go.sum ./
# Copy the utils module (required by the replace directive)
COPY utils/ /utils/

# Download dependencies
RUN go mod download

# Copy the source code
COPY schema-repository/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o schema-repository ./cmd

# Use a minimal alpine image for the final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/schema-repository .
# Copy config files
COPY --from=builder /app/config ./config

# Expose the port the service runs on
EXPOSE 8080

# Command to run the executable
CMD ["./schema-repository"]