# Schema Repository

A simple lightweight service for storing and retrieving JSON schemas with versioning support. Part of the go-feijoada
project.

## Overview

Schema Repository provides a RESTful API for managing JSON schemas with semantic versioning. It allows you to:

- Store JSON schemas with specific names and versions
- Retrieve schemas by name and version
- Delete schemas when they're no longer needed

The service uses Redis or Valkey as the backend storage system, making it fast and reliable.

## Features

- **Semantic Versioning**: Store multiple versions of the same schema
- **RESTful API**: Simple HTTP interface for schema management using [gin-gonic/gin](https://github.com/gin-gonic/gin)
- **Flexible Storage**: Support for both Redis and Valkey backends
- **Health Checks**: Built-in health check endpoints
  using [tavsec/gin-healthcheck](https://github.com/tavsec/gin-healthcheck)
- **Configurable**: Easy configuration via YAML files and environment variables
  using [knadh/koanf](https://github.com/knadh/koanf)

## Missing Features

- **Version compatibility check**: Schemas are not checked for retro-compatibility
- **Valkey integrated test**: Redis only

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Redis or Valkey instance

### Installation

```bash
# Clone the repository
git clone https://github.com/mfelipe/go-feijoada.git
cd go-feijoada/schema-repository

# Build the service
go build -o schema-repository ./cmd
```

### Configuration

Configuration is managed through YAML files and environment variables. The default configuration file is located at
`config/base.yaml`.

```yaml
sr:
  port: 8080
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
./schema-repository
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
  -d '{"type": "object", "properties": {"name": {"type": "string"}}}'
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
GET /health
```

## Integration with Other Services

Schema Repository is designed to work with other components in the go-feijoada project:

- **Schema Validator**: For validating data against stored schemas
- **Kafka Consumer**: Indirectly, for processing messages that requires schema validation

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

- `cmd/`: Application entry point
- `config/`: Configuration files and loading logic
- `internal/`: Internal packages
    - `clients/`: Redis and Valkey client implementations
    - `handlers/`: HTTP request handlers
    - `models/`: Data models
    - `repository/`: Storage layer abstraction
    - `service/`: Business logic

## License

See the [LICENSE](../LICENSE.md) file for details.