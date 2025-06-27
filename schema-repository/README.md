# Schema Repository

A simple service for storing and retrieving JSON schemas with versioning support. Part of the go-feijoada project.

## Overview

Schema Repository provides a RESTful API for managing JSON schemas with semantic versioning. It allows you to:

- Store JSON schemas with specific names and versions
- Retrieve schemas by name and version
- Delete schemas when they're no longer needed

The service uses Redis or Valkey as repository for low-latency operation and low resource consumption for this scenario.

## Features

- **Semantic Versioning**: Store multiple versions of the same schema
- **RESTful API**: Simple HTTP interface for schema management using [gin-gonic/gin](https://github.com/gin-gonic/gin)
- **Flexible Repository**: Support for Redis and Valkey backends (not using Valkey compatible Redis client for both)
- **Health Checks**: Built-in health check endpoints
  using [tavsec/gin-healthcheck](https://github.com/tavsec/gin-healthcheck)
- **Configurable**: Easy configuration via YAML files and environment variables
  using [knadh/koanf](https://github.com/knadh/koanf)

## Missing Features

- **Version compatibility check**: Schemas are not checked for retro-compatibility
- **Weak versioning**: Versions of a schema can be overwritten, which could cause invalid payloads to be valid and vice
  versa
- **Valkey integrated test**: Redis only

## Things that would be nice but may be out of the scope:

- Change Add Schema operation to:
  - Automatically increment the version accordingly to the input
  - Validate if the new schema version is compatible with the previous (if new MINOR or PATCH version)
- List Schema names and versions

## Usage Instructions
### Prerequisites
- Go 1.24 or higher
- Redis or Valkey instance

### Installation

```bash
# Clone the repository
git clone https://github.com/mfelipe/go-feijoada.git
```

### Configuration

Configuration is managed through YAML files and environment variables. The default configuration file is located at
`config/base.yaml`.

```yaml
sr:
  port: 8080
  log:
    level: "info"
  repository:
    data:
      keyPrefix: "schema-repository"
      keySeparator: ":"
```

You can configure Redis or Valkey connection details through environment variables:

```bash
# For Redis
export SR_REPOSITORY_REDIS_ADDRESS=localhost:6379
export SR_REPOSITORY_REDIS_PASSWORD=your_password

# For Valkey
export SR_REPOSITORY_VALKEY_ADDRESS=localhost:6379
export SR_REPOSITORY_VALKEY_PASSWORD=your_password
```

### Running the Service

```bash
# With running repository
go mod tidy
go build ./...
./schema-repository

# Composed with a repository
docker compose -f docker-compose.yml -p go-feijoada up -d
```

The service will start on the configured port (default: 8080).

## API Reference

### Create a Schema

```
POST /schemas/{name}/{version}
```

- `name`: Schema name
- `version`: Schema version (must follow semantic versioning format)

Request body should contain the JSON schema.

Example:

```bash
curl -X POST http://localhost:8080/schemas/user/1.0.0 \
  -H "Content-Type: application/json" \
  -d '{"schema":{"type": "object", "properties": {"name": {"type": "string"}}}}'
```

### Get a Schema

```
GET /schemas/{name}/{version}
```

- `name`: Schema name
- `version`: Schema version

Example:

```bash
curl http://localhost:8080/schemas/user/1.0.0
```

### Delete a Schema

```
DELETE /schemas/{name}/{version}
```

- `name`: Schema name
- `version`: Schema version

Example:

```bash
curl -X DELETE http://localhost:8080/schemas/user/1.0.0
```

## Health Check

The service includes a health check endpoint:

```
GET /healthz
```

## Integration with Other Services

Schema Repository is designed to work with other components in the go-feijoada project:

- **Schema Validator**: For validating data against stored schemas
- **Kafka Consumer**: Indirectly, for processing messages that requires schema validation

### Project Structure

- `cmd/`: Application entry point
- `config/`:
- `internal/`: Internal packages
    - `clients/`: Redis and Valkey client implementations
    - `handlers/`: HTTP request handlers
    - `models/`: Data models
    - `repository/`: Storage layer abstraction
    - `service/`: Business logic
